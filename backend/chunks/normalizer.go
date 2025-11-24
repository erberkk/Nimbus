package chunks

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// TextNormalizer handles text normalization and cleanup
type TextNormalizer struct {
	// Configuration flags
	removeExcessiveWhitespace bool
	normalizeLineEndings      bool
	stripLayoutArtifacts      bool
	preserveLists             bool
	preserveHeaders           bool
	maxConsecutiveNewlines    int
}

// NormalizerConfig defines configuration for text normalization
type NormalizerConfig struct {
	RemoveExcessiveWhitespace bool // Remove extra spaces and tabs
	NormalizeLineEndings      bool // Normalize to \n
	StripLayoutArtifacts      bool // Remove page numbers, headers, footers
	PreserveLists             bool // Keep list formatting
	PreserveHeaders           bool // Keep header formatting
	MaxConsecutiveNewlines    int  // Max blank lines to keep (default 2)
}

// DefaultNormalizerConfig returns default normalization configuration
func DefaultNormalizerConfig() NormalizerConfig {
	return NormalizerConfig{
		RemoveExcessiveWhitespace: true,
		NormalizeLineEndings:      true,
		StripLayoutArtifacts:      true,
		PreserveLists:             true,
		PreserveHeaders:           true,
		MaxConsecutiveNewlines:    2,
	}
}

// NewTextNormalizer creates a new text normalizer
func NewTextNormalizer(config NormalizerConfig) *TextNormalizer {
	if config.MaxConsecutiveNewlines <= 0 {
		config.MaxConsecutiveNewlines = 2
	}

	return &TextNormalizer{
		removeExcessiveWhitespace: config.RemoveExcessiveWhitespace,
		normalizeLineEndings:      config.NormalizeLineEndings,
		stripLayoutArtifacts:      config.StripLayoutArtifacts,
		preserveLists:             config.PreserveLists,
		preserveHeaders:           config.PreserveHeaders,
		maxConsecutiveNewlines:    config.MaxConsecutiveNewlines,
	}
}

// Normalize applies all normalization steps to text
func (n *TextNormalizer) Normalize(text string) string {
	// Step 1: Normalize line endings
	if n.normalizeLineEndings {
		text = n.normalizeLineEndingsFunc(text)
	}

	// Step 2: Strip layout artifacts (page numbers, repeated headers, etc.)
	if n.stripLayoutArtifacts {
		text = n.stripArtifacts(text)
	}

	// Step 3: Remove excessive whitespace
	if n.removeExcessiveWhitespace {
		text = n.removeExcessiveWhitespaceFunc(text)
	}

	// Step 4: Limit consecutive newlines
	text = n.limitConsecutiveNewlines(text)

	// Step 5: Clean up specific patterns
	text = n.cleanupPatterns(text)

	// Step 6: Final trim
	text = strings.TrimSpace(text)

	return text
}

// normalizeLineEndingsFunc converts all line endings to \n
func (n *TextNormalizer) normalizeLineEndingsFunc(text string) string {
	// Replace Windows line endings (CRLF)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	// Replace old Mac line endings (CR)
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

// removeExcessiveWhitespaceFunc removes extra spaces and tabs
func (n *TextNormalizer) removeExcessiveWhitespaceFunc(text string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		// Remove tabs
		line = strings.ReplaceAll(line, "\t", " ")

		// Preserve list markers
		if n.preserveLists && isListLine(line) {
			// Keep structure but normalize spacing after marker
			listRegex := regexp.MustCompile(`^(\s*[-*•]\s+|\s*\d+\.\s+)`)
			if match := listRegex.FindString(line); match != "" {
				rest := strings.TrimLeft(line[len(match):], " ")
				rest = regexp.MustCompile(`\s+`).ReplaceAllString(rest, " ")
				lines[i] = match + rest
				continue
			}
		}

		// Replace multiple spaces with single space
		line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")

		// Trim spaces from line (but preserve indentation if needed)
		lines[i] = strings.TrimSpace(line)
	}

	return strings.Join(lines, "\n")
}

// stripArtifacts removes common layout artifacts from PDF extraction
func (n *TextNormalizer) stripArtifacts(text string) string {
	lines := strings.Split(text, "\n")
	cleanedLines := []string{}

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines (will be handled later)
		if len(line) == 0 {
			cleanedLines = append(cleanedLines, line)
			continue
		}

		// Remove page numbers (standalone numbers or "Page X")
		if isPageNumber(line) {
			continue
		}

		// Remove common header/footer patterns
		if isHeaderFooter(line) {
			continue
		}

		// Remove lines with only special characters or dots (table of contents dots)
		if isOnlySpecialChars(line) {
			continue
		}

		// Remove repeated header lines (same line appears multiple times)
		if i > 0 && i < len(lines)-1 {
			if isRepeatedHeader(line, lines, i) {
				continue
			}
		}

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}

// limitConsecutiveNewlines reduces multiple blank lines to max allowed
func (n *TextNormalizer) limitConsecutiveNewlines(text string) string {
	// Build regex pattern for (n+1) or more consecutive newlines
	pattern := fmt.Sprintf(`\n{%d,}`, n.maxConsecutiveNewlines+1)
	regex := regexp.MustCompile(pattern)

	// Replace with exactly maxConsecutiveNewlines newlines
	replacement := strings.Repeat("\n", n.maxConsecutiveNewlines)

	return regex.ReplaceAllString(text, replacement)
}

// cleanupPatterns handles specific cleanup patterns
func (n *TextNormalizer) cleanupPatterns(text string) string {
	// Remove soft hyphens and zero-width characters
	text = strings.ReplaceAll(text, "\u00AD", "") // Soft hyphen
	text = strings.ReplaceAll(text, "\u200B", "") // Zero-width space
	text = strings.ReplaceAll(text, "\u200C", "") // Zero-width non-joiner
	text = strings.ReplaceAll(text, "\u200D", "") // Zero-width joiner
	text = strings.ReplaceAll(text, "\uFEFF", "") // Zero-width no-break space (BOM)

	// Fix hyphenated words split across lines
	text = regexp.MustCompile(`-\s*\n\s*`).ReplaceAllString(text, "")

	// Remove excessive punctuation repetition (but keep ellipsis)
	text = regexp.MustCompile(`\.{4,}`).ReplaceAllString(text, "...")
	text = regexp.MustCompile(`!{2,}`).ReplaceAllString(text, "!")
	text = regexp.MustCompile(`\?{2,}`).ReplaceAllString(text, "?")

	return text
}

// isListLine checks if a line is part of a list
func isListLine(line string) bool {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return false
	}

	// Check for bullet points
	if matched, _ := regexp.MatchString(`^\s*[-*•]\s+`, line); matched {
		return true
	}

	// Check for numbered lists
	if matched, _ := regexp.MatchString(`^\s*\d+\.\s+`, line); matched {
		return true
	}

	// Check for lettered lists
	if matched, _ := regexp.MatchString(`^\s*[a-z]\)\s+`, line); matched {
		return true
	}

	return false
}

// isPageNumber checks if a line is likely a page number
func isPageNumber(line string) bool {
	line = strings.TrimSpace(line)

	// Just a number
	if matched, _ := regexp.MatchString(`^\d+$`, line); matched {
		num := 0
		fmt.Sscanf(line, "%d", &num)
		// Likely page number if small-ish number
		return num < 10000
	}

	// "Page X" or "- X -" patterns
	pagePatterns := []string{
		`^[Pp]age\s+\d+$`,
		`^-\s*\d+\s*-$`,
		`^\d+\s*of\s*\d+$`,
	}

	for _, pattern := range pagePatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}

	return false
}

// isHeaderFooter checks if a line is likely a header or footer
func isHeaderFooter(line string) bool {
	line = strings.TrimSpace(line)

	// Very short lines at document edges are often headers/footers
	if len(line) < 5 {
		return false
	}

	// Common header/footer patterns
	patterns := []string{
		`^Copyright\s+©`,
		`^©\s+\d{4}`,
		`All rights reserved`,
		`^Confidential`,
		`^Draft`,
		`^\d{1,2}/\d{1,2}/\d{2,4}$`, // Date
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}

	return false
}

// isOnlySpecialChars checks if line contains only special characters
func isOnlySpecialChars(line string) bool {
	if len(strings.TrimSpace(line)) == 0 {
		return false
	}

	hasLetter := false
	for _, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasLetter = true
			break
		}
	}

	// If no letters or digits, it's only special chars
	return !hasLetter
}

// isRepeatedHeader checks if a line is repeated (common in PDFs with headers)
func isRepeatedHeader(line string, allLines []string, currentIndex int) bool {
	if len(line) > 100 {
		return false // Too long to be a repeated header
	}

	// Check if this exact line appears multiple times in document
	occurrences := 0
	for i, l := range allLines {
		if i == currentIndex {
			continue
		}
		if strings.TrimSpace(l) == line {
			occurrences++
		}
	}

	// If appears 3+ times, likely a repeated header
	return occurrences >= 3
}

// NormalizeForEmbedding performs minimal normalization suitable for embeddings
// This is lighter than full normalization to preserve semantic meaning
func NormalizeForEmbedding(text string) string {
	// Just normalize whitespace and line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)
	return text
}


