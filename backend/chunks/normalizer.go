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
	RemoveExcessiveWhitespace bool
	NormalizeLineEndings      bool
	StripLayoutArtifacts      bool
	PreserveLists             bool
	PreserveHeaders           bool
	MaxConsecutiveNewlines    int
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
	if n.normalizeLineEndings {
		text = n.normalizeLineEndingsFunc(text)
	}

	if n.stripLayoutArtifacts {
		text = n.stripArtifacts(text)
	}

	if n.removeExcessiveWhitespace {
		text = n.removeExcessiveWhitespaceFunc(text)
	}

	text = n.limitConsecutiveNewlines(text)
	text = n.cleanupPatterns(text)
	text = strings.TrimSpace(text)

	return text
}

// normalizeLineEndingsFunc converts all line endings to \n
func (n *TextNormalizer) normalizeLineEndingsFunc(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

// removeExcessiveWhitespaceFunc removes extra spaces and tabs
func (n *TextNormalizer) removeExcessiveWhitespaceFunc(text string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", " ")

		if n.preserveLists && isListLine(line) {
			listRegex := regexp.MustCompile(`^(\s*[-*•]\s+|\s*\d+\.\s+)`)
			if match := listRegex.FindString(line); match != "" {
				rest := strings.TrimLeft(line[len(match):], " ")
				rest = regexp.MustCompile(`\s+`).ReplaceAllString(rest, " ")
				lines[i] = match + rest
				continue
			}
		}

		line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")
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

		if len(line) == 0 {
			cleanedLines = append(cleanedLines, line)
			continue
		}

		if isPageNumber(line) {
			continue
		}

		if isHeaderFooter(line) {
			continue
		}

		if isOnlySpecialChars(line) {
			continue
		}

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
	pattern := fmt.Sprintf(`\n{%d,}`, n.maxConsecutiveNewlines+1)
	regex := regexp.MustCompile(pattern)

	replacement := strings.Repeat("\n", n.maxConsecutiveNewlines)

	return regex.ReplaceAllString(text, replacement)
}

// cleanupPatterns handles specific cleanup patterns
func (n *TextNormalizer) cleanupPatterns(text string) string {
	text = strings.ReplaceAll(text, "\u00AD", "") // Soft hyphen
	text = strings.ReplaceAll(text, "\u200B", "") // Zero-width space
	text = strings.ReplaceAll(text, "\u200C", "") // Zero-width non-joiner
	text = strings.ReplaceAll(text, "\u200D", "") // Zero-width joiner
	text = strings.ReplaceAll(text, "\uFEFF", "") // Zero-width no-break space (BOM)

	text = regexp.MustCompile(`-\s*\n\s*`).ReplaceAllString(text, "")

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

	if matched, _ := regexp.MatchString(`^\s*[-*•]\s+`, line); matched {
		return true
	}

	if matched, _ := regexp.MatchString(`^\s*\d+\.\s+`, line); matched {
		return true
	}

	if matched, _ := regexp.MatchString(`^\s*[a-z]\)\s+`, line); matched {
		return true
	}

	return false
}

// isPageNumber checks if a line is likely a page number
func isPageNumber(line string) bool {
	line = strings.TrimSpace(line)

	if matched, _ := regexp.MatchString(`^\d+$`, line); matched {
		num := 0
		fmt.Sscanf(line, "%d", &num)
		return num < 10000
	}

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

	if len(line) < 5 {
		return false
	}

	patterns := []string{
		`^Copyright\s+©`,
		`^©\s+\d{4}`,
		`All rights reserved`,
		`^Confidential`,
		`^Draft`,
		`^\d{1,2}/\d{1,2}/\d{2,4}$`,
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

	return !hasLetter
}

// isRepeatedHeader checks if a line is repeated (common in PDFs with headers)
func isRepeatedHeader(line string, allLines []string, currentIndex int) bool {
	if len(line) > 100 {
		return false
	}

	occurrences := 0
	for i, l := range allLines {
		if i == currentIndex {
			continue
		}
		if strings.TrimSpace(l) == line {
			occurrences++
		}
	}

	return occurrences >= 3
}

// NormalizeForEmbedding performs minimal normalization suitable for embeddings
// This is lighter than full normalization to preserve semantic meaning
func NormalizeForEmbedding(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)
	return text
}
