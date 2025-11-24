package chunks

import (
	"regexp"
	"strings"
)

// ChunkerConfig defines configuration for semantic text chunking
type ChunkerConfig struct {
	TargetTokens       int     // Target chunk size in tokens (default 1000)
	OverlapPercent     float64 // Overlap percentage (default 0.15 = 15%)
	CharsPerToken      int     // Approximate characters per token (default 4)
	MaxChunkSize       int     // Maximum chunk size in characters (0 = no limit)
	MinChunkSize       int     // Minimum chunk size in characters (default 100)
	PreserveParagraphs bool    // Try to preserve paragraph boundaries (default true)
}

// Chunk represents a text chunk with metadata
type Chunk struct {
	Index     int                    // Chunk index
	Text      string                 // Chunk text
	Metadata  map[string]interface{} // Optional metadata
	StartChar int                    // Start position in original text
	EndChar   int                    // End position in original text
}

// DefaultChunkerConfig returns default configuration
func DefaultChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		TargetTokens:       1000,
		OverlapPercent:     0.15,
		CharsPerToken:      4,
		MaxChunkSize:       0, // No limit
		MinChunkSize:       100,
		PreserveParagraphs: true,
	}
}

// SemanticTextSplitter handles intelligent text chunking
type SemanticTextSplitter struct {
	config ChunkerConfig
}

// NewSemanticTextSplitter creates a new text splitter with the given configuration
func NewSemanticTextSplitter(config ChunkerConfig) *SemanticTextSplitter {
	if config.TargetTokens <= 0 {
		config.TargetTokens = 1000
	}
	if config.CharsPerToken <= 0 {
		config.CharsPerToken = 4
	}
	if config.OverlapPercent < 0 {
		config.OverlapPercent = 0
	}
	if config.OverlapPercent > 0.5 {
		config.OverlapPercent = 0.5 // Cap at 50% overlap
	}
	if config.MinChunkSize <= 0 {
		config.MinChunkSize = 100
	}

	return &SemanticTextSplitter{
		config: config,
	}
}

// SplitSegments splits pre-processed segments into chunks
func (s *SemanticTextSplitter) SplitSegments(segments []TextSegment) []Chunk {
	var allChunks []Chunk
	chunkIndex := 0

	for _, segment := range segments {
		if segment.IsTable {
			// Tables get their own chunk - no splitting
			allChunks = append(allChunks, Chunk{
				Index:     chunkIndex,
				Text:      segment.Text,
				StartChar: segment.StartChar,
				EndChar:   segment.EndChar,
				Metadata:  s.extractChunkMetadata(segment.Text),
			})
			chunkIndex++
		} else {
			// Non-table text: split normally
			chunks := s.splitTextSegment(segment.Text, segment.StartChar)
			for i := range chunks {
				chunks[i].Index = chunkIndex
				chunkIndex++
			}
			allChunks = append(allChunks, chunks...)
		}
	}

	return allChunks
}

// Split splits text into semantic chunks (Legacy support, assumes no tables)
func (s *SemanticTextSplitter) Split(text string) []Chunk {
	// Treat as single text segment
	return s.SplitSegments([]TextSegment{{
		Text:      text,
		StartChar: 0,
		EndChar:   len(text),
		IsTable:   false,
	}})
}

// splitTextSegment splits a non-table text segment into chunks
func (s *SemanticTextSplitter) splitTextSegment(text string, baseOffset int) []Chunk {
	targetChars := s.config.TargetTokens * s.config.CharsPerToken
	overlapChars := int(float64(targetChars) * s.config.OverlapPercent)

	// Split into semantic units (paragraphs, then sentences)
	var units []textUnit
	if s.config.PreserveParagraphs {
		units = s.splitIntoParagraphs(text)
	} else {
		units = s.splitIntoSentences(text)
	}

	var chunks []Chunk
	var currentChunk strings.Builder
	var currentUnits []textUnit
	var currentStart int

	for i, unit := range units {
		// Calculate potential length if we add this unit
		potentialLength := currentChunk.Len() + len(unit.text)

		// Check if we should start a new chunk
		shouldSplit := false
		if potentialLength > targetChars && currentChunk.Len() > 0 {
			shouldSplit = true
		}

		// Also split if we exceed max chunk size (if configured)
		if s.config.MaxChunkSize > 0 && potentialLength > s.config.MaxChunkSize && currentChunk.Len() > 0 {
			shouldSplit = true
		}

		if shouldSplit {
			// Create chunk from current content
			chunkText := strings.TrimSpace(currentChunk.String())
			if len(chunkText) >= s.config.MinChunkSize {
				chunks = append(chunks, Chunk{
					Index:     0, // Will be set by caller
					Text:      chunkText,
					StartChar: baseOffset + currentStart,
					EndChar:   baseOffset + currentStart + len(chunkText),
					Metadata:  s.extractChunkMetadata(chunkText),
				})
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			currentUnits = []textUnit{}

			// Add overlap from previous chunk
			overlapSize := 0
			overlapUnits := []textUnit{}
			for j := len(currentUnits) - 1; j >= 0 && overlapSize < overlapChars; j-- {
				overlapSize += len(currentUnits[j].text)
				overlapUnits = append([]textUnit{currentUnits[j]}, overlapUnits...)
			}

			// Rebuild chunk with overlap
			for _, u := range overlapUnits {
				currentChunk.WriteString(u.text)
				currentChunk.WriteString(" ")
			}
			currentUnits = overlapUnits
			currentStart = units[i].start
		}

		// Add current unit
		if i > 0 && currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(unit.text)
		currentUnits = append(currentUnits, unit)

		if i == 0 {
			currentStart = unit.start
		}
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunkText := strings.TrimSpace(currentChunk.String())
		if len(chunkText) >= s.config.MinChunkSize {
			chunks = append(chunks, Chunk{
				Index:     0, // Will be set by caller
				Text:      chunkText,
				StartChar: baseOffset + currentStart,
				EndChar:   baseOffset + currentStart + len(chunkText),
				Metadata:  s.extractChunkMetadata(chunkText),
			})
		}
	}

	return chunks
}



// textUnit represents a semantic unit of text (sentence or paragraph)
type textUnit struct {
	text  string
	start int
	end   int
}

// splitIntoParagraphs splits text into paragraphs
func (s *SemanticTextSplitter) splitIntoParagraphs(text string) []textUnit {
	// Split by double newlines or more
	paragraphRegex := regexp.MustCompile(`\n\s*\n+`)
	parts := paragraphRegex.Split(text, -1)

	var units []textUnit
	currentPos := 0

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) == 0 {
			continue
		}

		// Find actual position in original text
		idx := strings.Index(text[currentPos:], trimmed)
		if idx >= 0 {
			start := currentPos + idx
			units = append(units, textUnit{
				text:  trimmed,
				start: start,
				end:   start + len(trimmed),
			})
			currentPos = start + len(trimmed)
		}
	}

	// If no paragraphs found or too few, fallback to sentences
	if len(units) < 2 {
		return s.splitIntoSentences(text)
	}

	return units
}

// splitIntoSentences splits text into sentences
func (s *SemanticTextSplitter) splitIntoSentences(text string) []textUnit {
	// Sentence boundary regex (punctuation followed by whitespace)
	// We manually check for uppercase letter after the match to simulate (?=[A-Z])
	sentenceRegex := regexp.MustCompile(`([.!?]+)\s+`)

	var units []textUnit
	lastEnd := 0

	matches := sentenceRegex.FindAllStringIndex(text, -1)

	for _, match := range matches {
		// match[0] is start of punctuation, match[1] is end of whitespace
		
		// Check if the character AFTER the match is an uppercase letter
		// This simulates the lookahead (?=[A-Z])
		isSentenceEnd := false
		if match[1] < len(text) {
			// We need to decode the rune to check if it's uppercase
			// But for simple ASCII check (which [A-Z] implies), we can just check byte range
			// or use unicode package if we want full support.
			// Let's use a simple check for now to avoid re-importing unicode if possible,
			// BUT we previously removed unicode. Let's re-add it to be safe and correct.
			nextChar := text[match[1]]
			if nextChar >= 'A' && nextChar <= 'Z' {
				isSentenceEnd = true
			}
		}

		if isSentenceEnd {
			sentenceEnd := match[1]
			// We include the punctuation and whitespace in the sentence for now, 
			// but TrimSpace will clean it up
			sentence := strings.TrimSpace(text[lastEnd:sentenceEnd])
			if len(sentence) > 0 {
				units = append(units, textUnit{
					text:  sentence,
					start: lastEnd,
					end:   sentenceEnd,
				})
			}
			lastEnd = sentenceEnd
		}
	}

	// Add final sentence
	if lastEnd < len(text) {
		sentence := strings.TrimSpace(text[lastEnd:])
		if len(sentence) > 0 {
			units = append(units, textUnit{
				text:  sentence,
				start: lastEnd,
				end:   len(text),
			})
		}
	}

	// Fallback: if no sentence boundaries found, split by length
	if len(units) == 0 {
		words := strings.Fields(text)
		targetWords := 50 // Approximately 50 words per unit

		for i := 0; i < len(words); i += targetWords {
			end := i + targetWords
			if end > len(words) {
				end = len(words)
			}
			sentence := strings.Join(words[i:end], " ")
			units = append(units, textUnit{
				text:  sentence,
				start: 0, // Position tracking not accurate for this fallback
				end:   0,
			})
		}
	}

	return units
}

// extractChunkMetadata extracts metadata from chunk text
func (s *SemanticTextSplitter) extractChunkMetadata(text string) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Estimate tokens only (useful for context window)
	metadata["estimated_tokens"] = len(text) / s.config.CharsPerToken

	return metadata
}


