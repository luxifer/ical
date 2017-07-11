package ical

import (
	"os"
	"testing"
)

var calendarList = []string{"fixtures/example.ics", "fixtures/with-alarm.ics"}

func TestParse(t *testing.T) {
	for _, filename := range calendarList {
		file, _ := os.Open(filename)
		_, err := Parse(file)
		file.Close()

		if err != nil {
			t.Error(err)
		}
	}
}

var dateList = []*Property{
	&Property{
		Name: "DTSTART",
		Params: map[string]*Param{
			"VALUE": &Param{
				Values: []string{"DATE"},
			},
		},
		Value: "19980119",
	},
	&Property{
		Name: "DTSTART",
		Params: map[string]*Param{
			"TZID": &Param{
				Values: []string{"America/New_York"},
			},
		},
		Value: "19980119T020000",
	},
	// Floating (local) date-time (no time zone)
	&Property{
		Name: "DSTART",
		Params: map[string]*Param{
			"VALUE": &Param{
				Values: []string{"DATE"},
			},
		},
		Value: "19980119T020000",
	},
	// Date-time in UTC
	&Property{
		Name: "DSTART",
		Params: map[string]*Param{
			"VALUE": &Param{
				Values: []string{"DATE"},
			},
		},
		Value: "19980119T070000Z",
	},
}

func TestParseDate(t *testing.T) {
	for _, prop := range dateList {
		out, err := parseDate(prop)

		if err != nil {
			t.Error(err)
		} else {
			t.Logf("in: %+v, out: %s", prop.Value, out)
		}
	}
}
