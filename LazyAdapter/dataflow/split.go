package dataflow

import (
	"strings"
)

type SplitFlow struct {
	source     DataFlow[string]
	delimiters string
	tokens     []string
	currentIdx int
}

func Split(delimiters string) func(DataFlow[string]) DataFlow[string] {
	return func(source DataFlow[string]) DataFlow[string] {
		return &SplitFlow{
			source:     source,
			delimiters: delimiters,
			currentIdx: -1,
		}
	}
}

func (s *SplitFlow) Next() bool {
	s.currentIdx++

	if s.tokens != nil && s.currentIdx < len(s.tokens) {
		return true
	}

	if !s.source.Next() {
		return false
	}

	content := s.source.Value()

	s.tokens = s.splitContent(content, s.delimiters)
	s.currentIdx = 0

	return len(s.tokens) > 0
}

func (s *SplitFlow) Value() string {
	if s.currentIdx < 0 || s.currentIdx >= len(s.tokens) {
		return ""
	}
	return s.tokens[s.currentIdx]
}

func (s *SplitFlow) Reset() {
	s.source.Reset()
	s.tokens = nil
	s.currentIdx = -1
}

func (s *SplitFlow) splitContent(content string, delimiters string) []string {

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
