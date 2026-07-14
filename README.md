# A11yLens

**Web Accessibility Analyzer CLI** — scan HTML files for common accessibility issues and get a score + grade.

A11yLens parses HTML and checks for WCAG-related issues including missing `alt` attributes, empty link text, missing input IDs, table header scope, iframe titles, and media captions.

## Features

- **🔍 HTML Parsing** — Full HTML5 document analysis using Go's `x/net/html` parser
- **📊 Scoring System** — Weighted scoring (0-100) with A-F letter grades
- **🔴 Severity Levels** — Issues categorized as critical, serious, moderate, or minor
- **📋 Multiple Output** — Terminal-friendly text or JSON for CI/CD pipelines
- **⚡ Fast** — Single-pass DOM analysis, no external dependencies beyond stdlib + x/net/html

## Checks

| Check | Severity | Description |
|-------|----------|-------------|
| `missing-alt` | Critical | `<img>` tags without `alt` attributes |
| `placeholder-link` | Serious | Links using `#` as href |
| `missing-input-id` | Serious | `<input>` fields without `id` (can't associate labels) |
| `missing-th-scope` | Moderate | `<th>` elements without `scope` attribute |
| `missing-iframe-title` | Serious | `<iframe>` elements without `title` attribute |
| `missing-video-captions` | Moderate | `<video>`/`<audio>` without `<track>` elements |

## Installation

```bash
# Download prebuilt binary
curl -sSfL https://github.com/EdgarOrtegaRamirez/a11ylens/releases/latest/download/a11ylens -o a11ylens
chmod +x a11ylens
sudo mv a11ylens /usr/local/bin/

# Or build from source
go install github.com/EdgarOrtegaRamirez/a11ylens@latest
```

## Usage

```bash
# Analyze an HTML file
a11ylens page.html

# Analyze a URL
a11ylens https://example.com

# JSON output for CI/CD
a11ylens page.html --json
```

## Terminal Output

```
============================================================
  A11yLens - Web Accessibility Checker
  URL: test.html
  Score: 40/100  Grade: F
============================================================

  Found 7 issue(s) across 23 elements

  🔴 CRITICAL (1)
    • Image missing alt attribute
      Element: <img src='photo.jpg'>

  🟠 SERIOUS (3)
    • Link uses '#' as href - likely a placeholder
    • Input field missing id attribute (cannot associate label)
    • iframe missing title attribute

  🟡 MODERATE (3)
    • Table header missing scope attribute (×2)
    • No <track> element found in video tag
```

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Run A11yLens
  run: |
    curl -sSfL https://github.com/EdgarOrtegaRamirez/a11ylens/releases/latest/download/a11ylens -o a11ylens
    chmod +x a11ylens
    ./a11ylens --json build/index.html > a11y-report.json
    score=$(jq '.score' a11y-report.json)
    if [ "$score" -lt 80 ]; then
      echo "Accessibility score below threshold: $score/100"
      exit 1
    fi
```

## License

MIT
