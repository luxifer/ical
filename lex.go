package ical

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos int      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	// Special tokens
	itemError itemType = iota
	itemEOF
	itemLineEnd

	// Literals
	itemName
	itemParamName
	itemParamValue
	itemValue

	// Misc
	itemColon     // :
	itemSemiColon // ;
	itemEqual     // =
	itemComma     // ,

	// Keyword
	itemKeyword // delimit the keyword list

	// Delimit
	itemBeginVCalendar
	itemEndVCalendar
	itemBeginVEvent
	itemEndVEvent
)

var key = map[string]itemType{
	"BEGIN:VCALENDAR": itemBeginVCalendar,
	"END:VCALENDAR":   itemEndVCalendar,
	"BEGIN:VEVENT":    itemBeginVEvent,
	"END:VEVENT":      itemEndVEvent,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input   string    // the string being scanned
	state   stateFn   // the next lexing function to enter
	start   int       // start position of this item
	pos     int       // current position in the input
	width   int       // width of last rune read from input
	lastPos int       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexName; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// State functions

const (
	crlf           = "\r\n"
	beginVCalendar = "BEGIN:VCALENDAR"
	endVCalendar   = "END:VCALENDAR"
	beginVEvent    = "BEGIN:VEVENT"
	endVEvent      = "END:VEVENT"
)

func lexContentLine(l *lexer) stateFn {
	switch r := l.next(); {
	case r == ';':
		l.emit(itemSemiColon)
		return lexParamName
	case r == ':':
		l.emit(itemColon)
		return lexValue
	case r == ',':
		l.emit(itemComma)
		return lexParamValue
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
}

// lexNewLine scans CRLF
func lexNewLine(l *lexer) stateFn {
	if l.peek() == eof {
		return nil
	}

	if !strings.HasPrefix(l.input[l.pos:], crlf) {
		l.errorf("unable to find end of line \"CRLF\"")
	}

	l.pos += len(crlf)
	l.emit(itemLineEnd)

	if l.next() == eof {
		l.emit(itemEOF)
		return nil
	}

	l.backup()

	return lexName
}

// lexName scans the name in the content line
//
// name       = iana-token / x-name
// iana-token = 1*(ALPHA / DIGIT / "-") ; iCalendar identifier registered with IANA
// x-name     = "X-" [vendorid "-"] 1*(ALPHA / DIGIT / "-") ; Reserved for experimental use.
// vendorid   = 3*(ALPHA / DIGIT) ; Vendor identification
func lexName(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], beginVCalendar) {
		l.pos += len(beginVCalendar)
		l.emit(itemBeginVCalendar)
		return lexNewLine
	}

	if strings.HasPrefix(l.input[l.pos:], endVCalendar) {
		l.pos += len(endVCalendar)
		l.emit(itemEndVCalendar)
		return lexNewLine
	}

	if strings.HasPrefix(l.input[l.pos:], beginVEvent) {
		l.pos += len(beginVEvent)
		l.emit(itemBeginVEvent)
		return lexNewLine
	}

	if strings.HasPrefix(l.input[l.pos:], endVEvent) {
		l.pos += len(endVEvent)
		l.emit(itemEndVEvent)
		return lexNewLine
	}

Loop:
	for {
		switch r := l.next(); {
		case isName(r):
			// absorb
		default:
			l.backup()
			l.emit(itemName)
			break Loop
		}
	}
	return lexContentLine
}

// lexParamName scans the param-name in the content line
//
// param-name = iana-token / x-name
// iana-token = 1*(ALPHA / DIGIT / "-") ; iCalendar identifier registered with IANA
// x-name     = "X-" [vendorid "-"] 1*(ALPHA / DIGIT / "-") ; Reserved for experimental use.
// vendorid   = 3*(ALPHA / DIGIT) ; Vendor identification
func lexParamName(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isName(r):
			// absorb
		default:
			l.backup()
			l.emit(itemParamName)
			break Loop
		}
	}

	r := l.next()

	if r == '=' {
		l.emit(itemEqual)
		return lexParamValue
	}
	return l.errorf("missing \"=\" sign after param name, got %#U", r)
}

// lexParamValue scans the param-value in the content line
//
// param-value   = paramtext / quoted-string
// paramtext     = *SAFE-CHAR
// quoted-string = DQUOTE *QSAFE-CHAR DQUOTE
// QSAFE-CHAR    = WSP / %x21 / %x23-7E / NON-US-ASCII ; Any character except CONTROL and DQUOTE
// SAFE-CHAR     = WSP / %x21 / %x23-2B / %x2D-39 / %x3C-7E / NON-US-ASCII ; Any character except CONTROL, DQUOTE, ";", ":", ","
func lexParamValue(l *lexer) stateFn {
	r := l.next()

	if r == '"' {
		l.ignore()
	QLoop:
		for {
			switch r := l.next(); {
			case isQSafeChar(r):
				// absorb
			default:
				l.backup()
				l.emit(itemParamValue)
				break QLoop
			}
		}

		r := l.next()

		if r != '"' {
			l.errorf("Missing \" for closing value")
		} else {
			l.ignore()
		}
	} else {
		l.backup()
	Loop:
		for {
			switch r := l.next(); {
			case isSafeChar(r):
				// absorb
			default:
				l.backup()
				l.emit(itemParamValue)
				break Loop
			}
		}
	}

	return lexContentLine
}

// lexValue scans the value in the content line
//
// value      = *VALUE-CHAR
// VALUE-CHAR = WSP / %x21-7E / NON-US-ASCII ; Any textual character
func lexValue(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case unicode.IsGraphic(r):
			// absorb
		default:
			l.backup()
			l.emit(itemValue)
			break Loop
		}
	}

	return lexNewLine
}

// rune helpers

func isName(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
}

func isQSafeChar(r rune) bool {
	return !unicode.IsControl(r) && r != '"'
}

func isSafeChar(r rune) bool {
	return !unicode.IsControl(r) && r != '"' && r != ';' && r != ':' && r != ','
}

// item helpers

// isItemName checks if the item is an ical name
func isItemName(i item) bool {
	return i.typ == itemName
}
