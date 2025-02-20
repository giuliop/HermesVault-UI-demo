package models

import "github.com/giuliop/HermesVault-frontend/config"

// Address represents a valid Algorand address
type Address string

func (a Address) Start() string {
	return splitEnds(string(a), config.NumCharsToHighlight, Start)
}
func (a Address) Middle() string {
	return splitEnds(string(a), config.NumCharsToHighlight, Middle)
}
func (a Address) End() string {
	return splitEnds(string(a), config.NumCharsToHighlight, End)
}

type part int

const (
	Start = iota
	Middle
	End
)

// splitEnds splits a string into three parts: the first n characters, the middle
// part, the last n characters
func splitEnds(s string, n int, part part) string {
	var start, middle, end string
	if len(s) <= 2*n {
		start, middle, end = s, "", ""
	} else {
		start, middle, end = s[:n], s[n:len(s)-n], s[len(s)-n:]
	}
	if part == Start {
		return start
	} else if part == Middle {
		return middle
	} else {
		return end
	}
}
