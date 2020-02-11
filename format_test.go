package ical

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	event := NewEvent()
	event.UID = "123@example.org"
	event.Timestamp = time.Date(2020, 2, 11, 0, 0, 0, 0, time.UTC)
	event.Summary = "Test event"

	cal := NewCalendar()
	cal.Events = []*Event{event}
	cal.Prodid = "-//ABC Corporation//NONSGML My Product//EN"
	cal.Version = "2.0"

	want := `BEGIN:VCALENDAR
PRODID:-//ABC Corporation//NONSGML My Product//EN
VERSION:2.0
CALSCALE:GREGORIAN
BEGIN:VEVENT
UID:123@example.org
DTSTAMP:20200211T000000Z
SUMMARY:Test event
END:VEVENT
END:VCALENDAR
`
	want = strings.Replace(want, "\n", "\r\n", -1)

	var buf bytes.Buffer
	if err := Format(&buf, cal); err != nil {
		t.Fatalf("Format() = %v", err)
	}

	if s := buf.String(); s != want {
		t.Errorf("Format() = \n%v\n but want \n%v", s, want)
	}
}
