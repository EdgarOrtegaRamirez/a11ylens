package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// A11yIssue represents a single accessibility issue found during analysis.
type A11yIssue struct {
	Severity string `json:"severity"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Element  string `json:"element,omitempty"`
}

// A11yReport is the complete accessibility analysis report.
type A11yReport struct {
	URL         string      `json:"url"`
	Issues      []A11yIssue `json:"issues"`
	Score       float64     `json:"score"`
	Grade       string      `json:"grade"`
	TotalChecks int         `json:"total_checks"`
	Passed      int         `json:"passed"`
	Failed      int         `json:"failed"`
	Warnings    int         `json:"warnings"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <URL|file> [--json]\n", os.Args[0])
		os.Exit(1)
	}

	target := os.Args[1]
	jsonOutput := false
	for _, arg := range os.Args[2:] {
		if arg == "--json" {
			jsonOutput = true
		}
	}

	var report *A11yReport
	var err error

	// Check if it's a file path or a URL
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		report, err = analyzeURL(target)
	} else {
		// Treat as file path
		content, readErr := os.ReadFile(target)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", readErr)
			os.Exit(1)
		}
		report, err = analyzeHTML(string(content), target)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report)
	} else {
		printTerminal(report)
	}
}

// analyzeHTML analyzes raw HTML content from a string.
func analyzeHTML(htmlContent string, baseURL string) (*A11yReport, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	var issues []A11yIssue
	totalChecks := 0

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		token := tokenizer.Token()
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			issues = checkElement(token, issues)
			totalChecks++
		}
	}

	report := &A11yReport{
		URL:         baseURL,
		Issues:      issues,
		TotalChecks: totalChecks,
	}

	report.Failed = 0
	report.Warnings = 0
	report.Passed = totalChecks
	for _, issue := range issues {
		if issue.Severity == "critical" || issue.Severity == "serious" {
			report.Failed++
			report.Passed--
		} else {
			report.Warnings++
			report.Passed--
		}
	}

	report.Score = calculateScore(report)
	report.Grade = scoreToGrade(report.Score)

	return report, nil
}

func analyzeURL(targetURL string) (*A11yReport, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure scheme
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	client := &http.Client{
		Timeout: 30,
	}

	resp, err := client.Get(parsedURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, parsedURL.String())
	}

	tokenizer := html.NewTokenizer(resp.Body)
	var issues []A11yIssue
	totalChecks := 0

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		token := tokenizer.Token()
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			issues = checkElement(token, issues)
			totalChecks++
		}
	}

	report := &A11yReport{
		URL:         parsedURL.String(),
		Issues:      issues,
		TotalChecks: totalChecks,
	}

	report.Failed = 0
	report.Warnings = 0
	report.Passed = totalChecks
	for _, issue := range issues {
		if issue.Severity == "critical" || issue.Severity == "serious" {
			report.Failed++
			report.Passed--
		} else {
			report.Warnings++
			report.Passed--
		}
	}

	report.Score = calculateScore(report)
	report.Grade = scoreToGrade(report.Score)

	return report, nil
}

func checkElement(token html.Token, issues []A11yIssue) []A11yIssue {
	tag := token.Data

	switch tag {
	case "img":
		issues = checkImgAlt(token, issues)
	case "a":
		issues = checkLinkText(token, issues)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		issues = checkHeading(token, issues)
	case "input":
		issues = checkInput(token, issues)
	case "th":
		issues = checkTh(token, issues)
	case "video", "audio":
		issues = checkMedia(token, issues)
	case "iframe":
		issues = checkIframe(token, issues)
	case "button":
		issues = checkButtonText(token, issues)
	}

	return issues
}

func checkImgAlt(token html.Token, issues []A11yIssue) []A11yIssue {
	var hasAlt bool
	for _, attr := range token.Attr {
		if attr.Key == "alt" {
			hasAlt = true
		}
	}
	if !hasAlt {
		src := getImageSrc(token)
		issues = append(issues, A11yIssue{
			Severity: "critical",
			Type:     "missing-alt",
			Message:  "Image missing alt attribute",
			Element:  fmt.Sprintf("<img src='%s'>", src),
		})
	}
	return issues
}

func getImageSrc(token html.Token) string {
	for _, attr := range token.Attr {
		if attr.Key == "src" {
			return attr.Val
		}
	}
	return ""
}

func checkLinkText(token html.Token, issues []A11yIssue) []A11yIssue {
	for _, attr := range token.Attr {
		if attr.Key == "href" && attr.Val == "#" {
			issues = append(issues, A11yIssue{
				Severity: "serious",
				Type:     "placeholder-link",
				Message:  "Link uses '#' as href - likely a placeholder",
				Element:  "<a href='#'>...</a>",
			})
			break
		}
	}
	return issues
}

func checkHeading(token html.Token, issues []A11yIssue) []A11yIssue {
	// Heading tags are checked for proper nesting in a full implementation.
	// For now, we just note that headings were found.
	_ = token
	return issues
}

func checkInput(token html.Token, issues []A11yIssue) []A11yIssue {
	var inputType string
	var hasID bool
	for _, attr := range token.Attr {
		if attr.Key == "type" {
			inputType = strings.ToLower(attr.Val)
		}
		if attr.Key == "id" {
			hasID = true
		}
	}

	// Skip buttons and submit
	if inputType == "submit" || inputType == "button" || inputType == "hidden" || inputType == "" {
		return issues
	}

	if !hasID {
		issues = append(issues, A11yIssue{
			Severity: "serious",
			Type:     "missing-input-id",
			Message:  "Input field missing id attribute (cannot associate label)",
			Element:  fmt.Sprintf("<input type='%s'>", inputType),
		})
	}

	return issues
}

func checkTh(token html.Token, issues []A11yIssue) []A11yIssue {
	var scope string
	for _, attr := range token.Attr {
		if attr.Key == "scope" {
			scope = attr.Val
		}
	}
	if scope == "" {
		issues = append(issues, A11yIssue{
			Severity: "moderate",
			Type:     "missing-th-scope",
			Message:  "Table header missing scope attribute",
			Element:  "<th>",
		})
	}
	return issues
}

func checkMedia(token html.Token, issues []A11yIssue) []A11yIssue {
	tag := token.Data
	issues = append(issues, A11yIssue{
		Severity: "moderate",
		Type:     fmt.Sprintf("missing-%s-captions", tag),
		Message:  fmt.Sprintf("No <track> element found in %s tag (no captions/subtitles)", tag),
		Element:  fmt.Sprintf("<%s>", tag),
	})
	return issues
}

func checkIframe(token html.Token, issues []A11yIssue) []A11yIssue {
	var hasTitle bool
	for _, attr := range token.Attr {
		if attr.Key == "title" {
			hasTitle = true
		}
	}
	if !hasTitle {
		issues = append(issues, A11yIssue{
			Severity: "serious",
			Type:     "missing-iframe-title",
			Message:  "iframe missing title attribute",
			Element:  "<iframe>",
		})
	}
	return issues
}

func checkButtonText(token html.Token, issues []A11yIssue) []A11yIssue {
	// Check for empty button text
	hasText := false
	for _, attr := range token.Attr {
		if attr.Key == "type" && attr.Val == "submit" {
			hasText = true // submit buttons can have value attribute
		}
	}
	// We can't check inner text in a simple tokenizer pass,
	// so we skip this check for now.
	_ = hasText
	return issues
}

func calculateScore(report *A11yReport) float64 {
	if report.TotalChecks == 0 {
		return 100.0
	}

	penalty := 0.0
	for _, issue := range report.Issues {
		switch issue.Severity {
		case "critical":
			penalty += 15.0
		case "serious":
			penalty += 10.0
		case "moderate":
			penalty += 5.0
		case "minor":
			penalty += 2.0
		}
	}

	score := 100.0 - penalty
	if score < 0 {
		score = 0
	}
	return float64(int(score*10)) / 10
}

func scoreToGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func printTerminal(report *A11yReport) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  A11yLens - Web Accessibility Checker\n")
	fmt.Printf("  URL: %s\n", report.URL)
	fmt.Printf("  Score: %.0f/100  Grade: %s\n", report.Score, report.Grade)
	fmt.Println(strings.Repeat("=", 60))

	if len(report.Issues) == 0 {
		fmt.Println("\n  ✓ No accessibility issues found!")
		fmt.Printf("  Checked %d elements\n", report.TotalChecks)
		return
	}

	fmt.Printf("\n  Found %d issue(s) across %d elements\n\n", len(report.Issues), report.TotalChecks)

	// Group by severity
	severities := map[string][]A11yIssue{
		"critical": {},
		"serious":  {},
		"moderate": {},
		"minor":    {},
	}
	for _, issue := range report.Issues {
		severities[issue.Severity] = append(severities[issue.Severity], issue)
	}

	severityOrder := []string{"critical", "serious", "moderate", "minor"}
	for _, sev := range severityOrder {
		issues := severities[sev]
		if len(issues) == 0 {
			continue
		}

		emoji := "⚠️"
		switch sev {
		case "critical":
			emoji = "🔴"
		case "serious":
			emoji = "🟠"
		case "moderate":
			emoji = "🟡"
		case "minor":
			emoji = "🔵"
		}

		fmt.Printf("  %s %s (%d)\n", emoji, strings.ToUpper(sev), len(issues))
		for _, issue := range issues {
			fmt.Printf("    • %s\n", issue.Message)
			if issue.Element != "" {
				fmt.Printf("      Element: %s\n", issue.Element)
			}
		}
		fmt.Println()
	}
}
