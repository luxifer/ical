package ical

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type Calendar struct {
	Properties []*Property
	Events     []*Event
	Prodid     string
	Version    string
	Calscale   string
	Method     string
}

type Event struct {
	Properties  []*Property
	Uid         string
	Timestamp   time.Time
	StartDate   time.Time
	EndDate     time.Time
	Summary     string
	Description string
}

type Property struct {
	Name   string
	Params map[string]*Param
	Value  string
}

type Param struct {
	Values []string
}

type parser struct {
	lex       *lexer
	token     [2]item
	peekCount int
	scope     int
	c         *Calendar
	v         *Event
}

// Parse transforms the raw iCalendar into a Calendar struct
// It's to the caller to close the io.Reader
func Parse(r io.Reader) (*Calendar, error) {
	p := &parser{}
	p.c = NewCalendar()
	p.scope = scopeCalendar
	bytes, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	text := unfold(string(bytes))
	p.lex = lex(text)
	return p.parse()
}

// NewCalendar creates an empty Calendar
func NewCalendar() *Calendar {
	c := &Calendar{}
	c.Properties = make([]*Property, 0)
	c.Events = make([]*Event, 0)
	return c
}

// NewProperty creates an empty Property
func NewProperty() *Property {
	p := &Property{}
	p.Params = make(map[string]*Param)
	return p
}

// NewEvent creates an empty Event
func NewEvent() *Event {
	v := &Event{}
	v.Properties = make([]*Property, 0)
	return v
}

// NewParam creates an empty Param
func NewParam() *Param {
	p := &Param{}
	p.Values = make([]string, 0)
	return p
}

// unfold convert multiple line value to one line
func unfold(text string) string {
	return strings.Replace(text, "\r\n ", "", -1)
}

// next returns the next token.
func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *parser) backup() {
	p.peekCount++
}

// peek returns but does not consume the next token.
func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

// parse

const (
	scopeCalendar int = iota
	scopeEvent
)

const (
	dateLayout              = "20060102"
	dateTimeLayoutUTC       = "20060102T150405Z"
	dateTimeLayoutLocalized = "20060102T150405"
)

var done = errors.New("done")

func (p *parser) parse() (*Calendar, error) {
	if item := p.next(); item.typ != itemBeginVCalendar {
		return nil, fmt.Errorf("found %s, expected BEGIN:VCALENDAR", item)
	}

	if item := p.next(); item.typ != itemLineEnd {
		return nil, fmt.Errorf("found %s, expected CRLF", item)
	}

	for {
		err := p.scanContentLine()

		if err == done {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	return p.c, nil
}

// scanContentLine parses a content-line of a calendar
func (p *parser) scanContentLine() error {
	name := p.next()

	if name.typ == itemBeginVEvent {
		if err := validateCalendar(p.c); err != nil {
			return err
		}

		p.v = NewEvent()
		p.scope = scopeEvent

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}

		return p.scanContentLine()
	}

	if name.typ == itemEndVEvent {
		if err := validateEvent(p.v); err != nil {
			return err
		}

		p.c.Events = append(p.c.Events, p.v)
		p.scope = scopeCalendar

		if item := p.next(); item.typ != itemLineEnd {
			return fmt.Errorf("found %s, expected CRLF", item)
		}

		return p.scanContentLine()
	}

	if name.typ == itemEndVCalendar {
		if p.scope == scopeEvent {
			return fmt.Errorf("found %s, expeced END:VEVENT", name)
		} else {
			return done
		}
	}

	if !isItemName(name) {
		return fmt.Errorf("found %s, expected a \"name\" token", name)
	}

	prop := NewProperty()
	prop.Name = name.val

	if err := p.scanParams(prop); err != nil {
		return err
	}

	if item := p.next(); item.typ != itemColon {
		return fmt.Errorf("found %s, expected \":\"", item)
	}

	value := p.next()

	if value.typ != itemValue {
		fmt.Errorf("found %s, expected a value", value)
	}

	prop.Value = value.val

	if item := p.next(); item.typ != itemLineEnd {
		return fmt.Errorf("found %s, expected CRLF", name)
	}

	if p.scope == scopeCalendar {
		p.c.Properties = append(p.c.Properties, prop)
	} else {
		p.v.Properties = append(p.v.Properties, prop)
	}

	return nil
}

// scanParams parses a list of param inside a content-line
func (p *parser) scanParams(prop *Property) error {
	for {
		item := p.next()

		if item.typ != itemSemiColon {
			p.backup()
			return nil
		}

		paramName := p.next()

		if paramName.typ != itemParamName {
			return fmt.Errorf("found %s, expected a param-name", paramName)
		}

		param := NewParam()

		if item := p.next(); item.typ != itemEqual {
			return fmt.Errorf("found %s, expected =", item)
		}

		if err := p.scanValues(param); err != nil {
			return err
		}

		prop.Params[paramName.val] = param
	}
}

// scanValues parses a list of at least one value for a param
func (p *parser) scanValues(param *Param) error {
	paramValue := p.next()

	if paramValue.typ != itemParamValue {
		return fmt.Errorf("found %s, expected a param-value", paramValue)
	}

	param.Values = append(param.Values, paramValue.val)

	for {
		item := p.next()

		if item.typ != itemComma {
			p.backup()
			return nil
		}

		paramValue := p.next()

		if paramValue.typ != itemParamValue {
			return fmt.Errorf("found %s, expected a param-value", paramValue)
		}

		param.Values = append(param.Values, paramValue.val)
	}
}

// validateCalendar validate calendar props
func validateCalendar(c *Calendar) error {
	requiredCount := 0
	for _, prop := range c.Properties {
		if prop.Name == "PRODID" {
			c.Prodid = prop.Value
			requiredCount++
		}

		if prop.Name == "VERSION" {
			c.Version = prop.Value
			requiredCount++
		}

		if prop.Name == "CALSCALE" {
			c.Calscale = prop.Value
		}

		if prop.Name == "METHOD" {
			c.Method = prop.Value
		}
	}

	if requiredCount != 2 {
		return fmt.Errorf("missing either required property \"prodid / version /\"")
	}

	return nil
}

// validateEvent validate event props
func validateEvent(v *Event) error {
	requiredCount := 0
	uniqueCount := make(map[string]int)
	for _, prop := range v.Properties {
		if prop.Name == "UID" {
			v.Uid = prop.Value
			uniqueCount["UID"]++
			requiredCount++
		}

		if prop.Name == "DTSTAMP" {
			v.Timestamp, _ = time.Parse(dateTimeLayoutUTC, prop.Value)
			uniqueCount["DTSTAMP"]++
			requiredCount++
		}

		if prop.Name == "DTSTART" {
			v.StartDate, _ = parseDate(prop)
			uniqueCount["DTSTATR"]++
			requiredCount++
		}

		if prop.Name == "DTEND" {
			if hasProperty("DURATION", v.Properties) {
				return fmt.Errorf("Either \"dtend\" or \"duration\" MAY appear")
			}
			v.EndDate, _ = parseDate(prop)
			uniqueCount["DTEND"]++
		}

		if prop.Name == "DURATION" {
			if hasProperty("DTEND", v.Properties) {
				return fmt.Errorf("Either \"dtend\" or \"duration\" MAY appear")
			}
			uniqueCount["DURATION"]++
		}

		if prop.Name == "SUMMARY" {
			v.Summary = prop.Value
			uniqueCount["SUMMARY"]++
		}

		if prop.Name == "DESCRIPTION" {
			v.Description = prop.Value
			uniqueCount["DESCRIPTION"]++
		}
	}

	if requiredCount != 3 {
		return fmt.Errorf("missing either required property \"dtstamp / uid / dtstart /\"")
	}

	for key, value := range uniqueCount {
		if value > 1 {
			return fmt.Errorf("\"%s\" property must not occur more than once", key)
		}
	}

	if !hasProperty("DTEND", v.Properties) {
		v.EndDate = v.StartDate.Add(time.Hour * 24) // add one day to start date
	}

	return nil
}

// hasProperty checks if a given component has a certain property
func hasProperty(name string, properties []*Property) bool {
	for _, prop := range properties {
		if name == prop.Name {
			return true
		}
	}
	return false
}

// parseDate transform an ical date into a time.Time
func parseDate(prop *Property) (time.Time, error) {
	layout := dateTimeLayoutUTC
	if val, ok := prop.Params["VALUE"]; ok {
		switch val.Values[0] {
		case "DATE":
			layout = dateLayout
		}
	}
	if tz, ok := prop.Params["TZID"]; ok {
		loc, _ := time.LoadLocation(tz.Values[0])
		return time.ParseInLocation(dateTimeLayoutLocalized, prop.Value, loc)
	} else {
		return time.Parse(layout, prop.Value)
	}
}
