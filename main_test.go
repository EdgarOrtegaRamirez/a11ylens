package main

import (
	"testing"
)

func TestAnalyzeHTML(t *testing.T) {
	html := `<html><head><title>Test</title></head><body><h1>Test</h1><p>Content</p></body></html>`
	report, err := analyzeHTML(html, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.URL != "http://example.com" {
		t.Errorf("expected URL http://example.com, got %s", report.URL)
	}
	if report.TotalChecks == 0 {
		t.Error("expected some checks to run")
	}
}

func TestAnalyzeHTML_NoContent(t *testing.T) {
	html := `<html><head></head><body></body></html>`
	report, err := analyzeHTML(html, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 100.0 {
		t.Errorf("expected score 100.0 for empty content, got %f", report.Score)
	}
}

func TestAnalyzeHTML_MissingAlt(t *testing.T) {
	html := `<html><body><img src="test.jpg"></body></html>`
	report, err := analyzeHTML(html, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasMissingAlt := false
	for _, issue := range report.Issues {
		if issue.Type == "missing-alt" {
			hasMissingAlt = true
			break
		}
	}
	if !hasMissingAlt {
		t.Error("expected missing-alt issue for img without alt attribute")
	}
}

func TestAnalyzeHTML_MissingIframeTitle(t *testing.T) {
	html := `<html><body><iframe src="https://example.com"></body></html>`
	report, err := analyzeHTML(html, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasMissingTitle := false
	for _, issue := range report.Issues {
		if issue.Type == "missing-iframe-title" {
			hasMissingTitle = true
			break
		}
	}
	if !hasMissingTitle {
		t.Error("expected missing-iframe-title issue")
	}
}

func TestGrade(t *testing.T) {
	tests := []struct {
		score  float64
		expect string
	}{
		{100.0, "A"},
		{95.0, "A"},
		{90.0, "A"},
		{85.0, "B"},
		{80.0, "B"},
		{75.0, "C"},
		{70.0, "C"},
		{65.0, "D"},
		{60.0, "D"},
		{50.0, "F"},
		{40.0, "F"},
		{30.0, "F"},
		{0.0, "F"},
	}
	for _, tt := range tests {
		grade := scoreToGrade(tt.score)
		if grade != tt.expect {
			t.Errorf("score %f -> expected %s, got %s", tt.score, tt.expect, grade)
		}
	}
}

func TestPrintTerminal(t *testing.T) {
	report := &A11yReport{
		URL:         "http://example.com",
		Score:       85.0,
		Grade:       "B",
		TotalChecks: 10,
		Passed:      8,
		Failed:      1,
		Warnings:    1,
		Issues: []A11yIssue{
			{Severity: "warning", Type: "missing-alt", Message: "Image missing alt text"},
		},
	}
	// Just verify it doesn't panic
	printTerminal(report)
}
