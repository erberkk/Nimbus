package chunks

import (
	"regexp"
	"strings"
)

// ChunkerConfig defines configuration for semantic text chunking
type ChunkerConfig struct {
	TargetTokens       int
	OverlapPercent     float64
	CharsPerToken      int
	MaxChunkSize       int
	MinChunkSize       int
	PreserveParagraphs bool
}

// Chunk represents a text chunk with metadata
type Chunk struct {
	Index     int
	Text      string
	Metadata  map[string]interface{}
	StartChar int
	EndChar   int
}

// DefaultChunkerConfig returns default configuration
func DefaultChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		TargetTokens:       1000,
		OverlapPercent:     0.15,
		CharsPerToken:      4,
		MaxChunkSize:       0,
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
		config.OverlapPercent = 0.5
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
			allChunks = append(allChunks, Chunk{
				Index:     chunkIndex,
				Text:      segment.Text,
				StartChar: segment.StartChar,
				EndChar:   segment.EndChar,
				Metadata:  s.extractChunkMetadata(segment.Text),
			})
			chunkIndex++
		} else {
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
		potentialLength := currentChunk.Len() + len(unit.text)

		shouldSplit := false
		if potentialLength > targetChars && currentChunk.Len() > 0 {
			shouldSplit = true
		}

		if s.config.MaxChunkSize > 0 && potentialLength > s.config.MaxChunkSize && currentChunk.Len() > 0 {
			shouldSplit = true
		}

		if shouldSplit {
			chunkText := strings.TrimSpace(currentChunk.String())
			if len(chunkText) >= s.config.MinChunkSize {
				chunks = append(chunks, Chunk{
					Index:     0,
					Text:      chunkText,
					StartChar: baseOffset + currentStart,
					EndChar:   baseOffset + currentStart + len(chunkText),
					Metadata:  s.extractChunkMetadata(chunkText),
				})
			}

			currentChunk.Reset()
			currentUnits = []textUnit{}

			overlapSize := 0
			overlapUnits := []textUnit{}
			for j := len(currentUnits) - 1; j >= 0 && overlapSize < overlapChars; j-- {
				overlapSize += len(currentUnits[j].text)
				overlapUnits = append([]textUnit{currentUnits[j]}, overlapUnits...)
			}

			for _, u := range overlapUnits {
				currentChunk.WriteString(u.text)
				currentChunk.WriteString(" ")
			}
			currentUnits = overlapUnits
			currentStart = units[i].start
		}

		if i > 0 && currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(unit.text)
		currentUnits = append(currentUnits, unit)

		if i == 0 {
			currentStart = unit.start
		}
	}

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
	paragraphRegex := regexp.MustCompile(`\n\s*\n+`)
	parts := paragraphRegex.Split(text, -1)

	var units []textUnit
	currentPos := 0

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) == 0 {
			continue
		}

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
		isSentenceEnd := false
		if match[1] < len(text) {
			nextChar := text[match[1]]
			if nextChar >= 'A' && nextChar <= 'Z' {
				isSentenceEnd = true
			}
		}

		if isSentenceEnd {
			sentenceEnd := match[1]
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

	if len(units) == 0 {
		words := strings.Fields(text)
		targetWords := 50

		for i := 0; i < len(words); i += targetWords {
			end := i + targetWords
			if end > len(words) {
				end = len(words)
			}
			sentence := strings.Join(words[i:end], " ")
			units = append(units, textUnit{
				text:  sentence,
				start: 0,
				end:   0,
			})
		}
	}

	return units
}

// extractChunkMetadata extracts metadata from chunk text
func (s *SemanticTextSplitter) extractChunkMetadata(text string) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["estimated_tokens"] = len(text) / s.config.CharsPerToken

	return metadata
}
