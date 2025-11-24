package chunks

import (
	"fmt"
	"regexp"
	"strings"
)

// TableProcessor handles detection and restructuring of tables in text
type TableProcessor struct {
	// Configuration could go here
}

// NewTableProcessor creates a new table processor
func NewTableProcessor() *TableProcessor {
	return &TableProcessor{}
}

// TextSegment represents a segment of text (either table or regular text)
type TextSegment struct {
	Text      string
	StartChar int
	EndChar   int
	IsTable   bool
}

// Process analyzes text and returns segments (tables are restructured)
func (tp *TableProcessor) Process(text string) []TextSegment {
	lines := strings.Split(text, "\n")
	var segments []TextSegment

	// Build line positions in original text
	linePositions := make([]int, len(lines))
	currentPos := 0
	for i, line := range lines {
		linePositions[i] = currentPos
		currentPos += len(line) + 1 // +1 for newline
	}

	i := 0
	textStart := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Check if this line is a table title
		if tp.isTableTitle(line) {
			// Add non-table text before table (if any)
			tableStart := linePositions[i]
			if tableStart > textStart {
				nonTableText := text[textStart:tableStart]
				if strings.TrimSpace(nonTableText) != "" {
					segments = append(segments, TextSegment{
						Text:      nonTableText,
						StartChar: textStart,
						EndChar:   tableStart,
						IsTable:   false,
					})
				}
			}

			// Extract and restructure table
			tableText, nextLineIdx := tp.processTableSection(lines, i)
			
			// Calculate end position (approximate since we restructured)
			// We use the position of the line after the table in the original text
			tableEndPos := currentPos
			if nextLineIdx < len(lines) {
				tableEndPos = linePositions[nextLineIdx]
			}

			// Add table segment
			segments = append(segments, TextSegment{
				Text:      tableText,
				StartChar: tableStart,
				EndChar:   tableEndPos,
				IsTable:   true,
			})

			// Update position
			textStart = tableEndPos
			i = nextLineIdx
		} else {
			i++
		}
	}

	// Add remaining non-table text
	if textStart < len(text) {
		remainingText := text[textStart:]
		if strings.TrimSpace(remainingText) != "" {
			segments = append(segments, TextSegment{
				Text:      remainingText,
				StartChar: textStart,
				EndChar:   len(text),
				IsTable:   false,
			})
		}
	}

	// If no segments found, return entire text as one segment
	if len(segments) == 0 {
		segments = append(segments, TextSegment{
			Text:      text,
			StartChar: 0,
			EndChar:   len(text),
			IsTable:   false,
		})
	}

	return segments
}

// isTableTitle checks if a line is likely a table title
func (tp *TableProcessor) isTableTitle(line string) bool {
	lower := strings.ToLower(line)

	// Check for explicit table keywords (even if embedded in other text)
	if strings.Contains(lower, "comparison") ||
		strings.Contains(lower, "table") ||
		strings.Contains(lower, "overview") {
		// If it contains "# Comparison" pattern, definitely a table title
		if strings.Contains(line, "# Comparison") || strings.Contains(line, "# comparison") {
			return true
		}
		// Also check if it's a standalone comparison phrase
		if regexp.MustCompile(`(?i)(comparison|table)\s+of\s+`).MatchString(line) {
			return true
		}
	}

	// Check for patterns like "X vs Y", "X-Y-Z"
	if regexp.MustCompile(`\d+-\d+`).MatchString(line) {
		return true
	}

	if regexp.MustCompile(`\svs\.?\s|\sversus\s`).MatchString(lower) {
		return true
	}

	return false
}

// processTableSection processes a detected table section and returns restructured text
func (tp *TableProcessor) processTableSection(lines []string, startIdx int) (string, int) {
	if startIdx >= len(lines) {
		return "", startIdx
	}

	titleLine := strings.TrimSpace(lines[startIdx])

	// Extract table title
	title := titleLine
	if idx := strings.Index(titleLine, "# Comparison"); idx >= 0 {
		title = strings.TrimSpace(titleLine[idx+2:])
	} else if idx := strings.Index(strings.ToLower(titleLine), "comparison of"); idx >= 0 {
		title = strings.TrimSpace(titleLine[idx:])
	}

	// Clean title
	if dashIdx := strings.Index(title, " - "); dashIdx > 0 {
		beforeDash := title[:dashIdx]
		afterDash := title[dashIdx+3:]
		if len(afterDash) < 50 && !regexp.MustCompile(`\d+-\d+`).MatchString(afterDash) {
			title = strings.TrimSpace(beforeDash)
		}
	}
	if dashIdx := strings.Index(title, " – "); dashIdx > 0 {
		title = strings.TrimSpace(title[:dashIdx])
	}
	titleLines := strings.Split(title, "\n")
	title = strings.TrimSpace(titleLines[0])

	i := startIdx + 1

	// Determine number of columns
	columnCount := tp.extractColumnCount(title)
	if columnCount < 2 {
		columnCount = tp.detectColumnCount(lines, i)
	}

	if columnCount < 2 {
		// Not a valid table, return original line and continue
		return lines[startIdx], startIdx + 1
	}

	// Extract column names
	columnNames := tp.extractColumnNames(title, columnCount)

	// Build structured table output
	var tableBuilder strings.Builder
	tableBuilder.WriteString(fmt.Sprintf("COMPARISON TABLE: %s\n\n", title))
	tableBuilder.WriteString("This table compares: ")
	for idx, name := range columnNames {
		if idx > 0 {
			tableBuilder.WriteString(", ")
		}
		tableBuilder.WriteString(name)
	}
	tableBuilder.WriteString("\n\n")

	// Read rows
	rowCount := 0
	maxRows := 50

	for i < len(lines) && rowCount < maxRows {
		line := strings.TrimSpace(lines[i])

		if line == "" || tp.isTableTitle(line) {
			break
		}

		// Skip placeholder headers
		lowerLine := strings.ToLower(line)
		if lowerLine == "feature:" || (strings.HasSuffix(lowerLine, ":") && len(line) < 20) {
			i++
			// Skip values
			for j := 0; j < columnCount && i < len(lines); j++ {
				if strings.TrimSpace(lines[i]) != "" {
					i++
				} else {
					break
				}
			}
			continue
		}

		rowHeader := line
		i++

		values := []string{}
		for j := 0; j < columnCount && i < len(lines); j++ {
			valueLine := strings.TrimSpace(lines[i])
			if valueLine == "" || tp.isTableTitle(valueLine) {
				break
			}
			values = append(values, valueLine)
			i++
		}

		if len(values) == columnCount {
			formattedRow := tp.formatTableRow(rowHeader, values, columnNames)
			if formattedRow != "" {
				tableBuilder.WriteString(formattedRow)
				tableBuilder.WriteString("\n")
				rowCount++
			}
		} else {
			i -= len(values)
			break
		}
	}

	if rowCount > 0 {
		return tableBuilder.String(), i
	}

	return lines[startIdx], startIdx + 1
}

func (tp *TableProcessor) extractColumnCount(title string) int {
	numberPattern := regexp.MustCompile(`\d+`)
	numbers := numberPattern.FindAllString(title, -1)

	if len(numbers) >= 2 {
		return len(numbers)
	}

	if regexp.MustCompile(`\svs\.?\s|\sversus\s`).MatchString(strings.ToLower(title)) {
		return 2
	}

	return 0
}

func (tp *TableProcessor) detectColumnCount(lines []string, startIdx int) int {
	if startIdx >= len(lines) {
		return 0
	}

	i := startIdx
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	if i >= len(lines) {
		return 0
	}

	for testCount := 2; testCount <= 5; testCount++ {
		if tp.looksLikeTableWithColumns(lines, i, testCount, 2) {
			return testCount
		}
	}

	return 0
}

func (tp *TableProcessor) looksLikeTableWithColumns(lines []string, startIdx, columnCount, minRows int) bool {
	i := startIdx
	rowsFound := 0

	for rowsFound < minRows && i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}

		headerLine := strings.TrimSpace(lines[i])
		if len(headerLine) < 3 {
			return false
		}
		i++

		valuesFound := 0
		for valuesFound < columnCount && i < len(lines) {
			valueLine := strings.TrimSpace(lines[i])
			if valueLine == "" {
				break
			}
			valuesFound++
			i++
		}

		if valuesFound != columnCount {
			return false
		}

		rowsFound++
	}

	return rowsFound >= minRows
}

func (tp *TableProcessor) extractColumnNames(title string, columnCount int) []string {
	lower := strings.ToLower(title)

	if strings.Contains(lower, "wifi") || strings.Contains(lower, "wi-fi") {
		numbers := regexp.MustCompile(`\d+`).FindAllString(title, -1)
		if len(numbers) >= columnCount {
			names := make([]string, columnCount)
			for i := 0; i < columnCount; i++ {
				names[i] = fmt.Sprintf("Wi-Fi %s", numbers[i])
			}
			return names
		}
	}

	names := make([]string, columnCount)
	for i := 0; i < columnCount; i++ {
		names[i] = fmt.Sprintf("Column %d", i+1)
	}
	return names
}

func (tp *TableProcessor) formatTableRow(header string, values []string, columnNames []string) string {
	var builder strings.Builder

	if strings.TrimSpace(strings.ToLower(header)) == "feature:" {
		return ""
	}

	builder.WriteString(fmt.Sprintf("\n%s\n", strings.TrimSpace(header)))

	for i, value := range values {
		if i < len(columnNames) {
			itemName := columnNames[i]
			cleanValue := strings.TrimSpace(value)
			if strings.HasPrefix(cleanValue, itemName+":") {
				cleanValue = strings.TrimSpace(strings.TrimPrefix(cleanValue, itemName+":"))
			}
			builder.WriteString(fmt.Sprintf("  %s: %s\n", itemName, cleanValue))
		} else {
			builder.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(value)))
		}
	}

	return builder.String()
}
