package dataflow

import (
	"strings"
)

type FileContentSplitFlow struct {
	source      DataFlow[FileContent]
	delimiters  string
	tokens      []string
	currentFile string
	currentIdx  int
}

func FileContentSplit(delimiters string) func(DataFlow[FileContent]) DataFlow[string] {
	return func(source DataFlow[FileContent]) DataFlow[string] {
		return &FileContentSplitFlow{
			source:     source,
			delimiters: delimiters,
			currentIdx: -1,
		}
	}
}

func (s *FileContentSplitFlow) Next() bool {
	s.currentIdx++

	if s.tokens != nil && s.currentIdx < len(s.tokens) {
		return true
	}

	if !s.source.Next() {
		return false
	}

	fileContent := s.source.Value()
	s.currentFile = fileContent.Path

	s.tokens = s.splitContent(string(fileContent.Content), s.delimiters)
	s.currentIdx = 0

	return len(s.tokens) > 0
}

func (s *FileContentSplitFlow) Value() string {
	if s.currentIdx < 0 || s.currentIdx >= len(s.tokens) {
		return ""
	}
	return s.tokens[s.currentIdx]
}

func (s *FileContentSplitFlow) Reset() {
	s.source.Reset()
	s.tokens = nil
	s.currentIdx = -1
}

func (s *FileContentSplitFlow) splitContent(content string, delimiters string) []string {

	delimMap := make(map[rune]bool)
	for _, delimiter := range delimiters {
		delimMap[delimiter] = true
	}

	var tokens []string
	var token strings.Builder

	for _, char := range content {
		if delimMap[char] {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
		} else {
			token.WriteRune(char)
		}
	}

	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}

	return tokens
}
