package ical

import (
	"io/ioutil"
	"testing"
)

func TestLex(t *testing.T) {
	ical, _ := ioutil.ReadFile("fixtures/example.ics")
	lexer := lex(string(ical))

	for {
		item := lexer.nextItem()

		if item.typ == itemEOF {
			break
		}

		if item.typ == itemError {
			t.Error(item)
		}
	}
}
