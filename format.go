package ical

import (
	"bytes"
	"io"
	"strconv"
)

// Format writes the calendar to the provided io.Writer.
func Format(w io.Writer, cal *Calendar) error {
	var props []*Property
	if cal.Prodid != "" {
		props = append(props, &Property{Name: "PRODID", Value: cal.Prodid})
	}
	if cal.Version != "" {
		props = append(props, &Property{Name: "VERSION", Value: cal.Version})
	}
	if cal.Calscale != "" {
		props = append(props, &Property{Name: "CALSCALE", Value: cal.Calscale})
	}
	if cal.Method != "" {
		props = append(props, &Property{Name: "METHOD", Value: cal.Method})
	}
	cal.Properties = setProperties(cal.Properties, props)

	if _, err := io.WriteString(w, beginVCalendar+crlf); err != nil {
		return err
	}

	if err := formatPropertiesList(w, cal.Properties); err != nil {
		return err
	}

	for _, event := range cal.Events {
		if err := formatEvent(w, event); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, endVCalendar+crlf)
	return err
}

func formatEvent(w io.Writer, event *Event) error {
	var props []*Property
	if event.UID != "" {
		props = append(props, &Property{Name: "UID", Value: event.UID})
	}
	// TODO: add TZID if necessary
	if !event.Timestamp.IsZero() {
		props = append(props, &Property{
			Name:  "DTSTAMP",
			Value: event.Timestamp.Format(dateTimeLayoutUTC),
		})
	}
	if !event.StartDate.IsZero() {
		props = append(props, &Property{
			Name:  "DTSTART",
			Value: event.StartDate.Format(dateTimeLayoutUTC),
		})
	}
	if !event.EndDate.IsZero() {
		props = append(props, &Property{
			Name:  "DTEND",
			Value: event.EndDate.Format(dateTimeLayoutUTC),
		})
	}
	if event.Summary != "" {
		props = append(props, &Property{Name: "SUMMARY", Value: event.Summary})
	}
	if event.Description != "" {
		props = append(props, &Property{Name: "DESCRIPTION", Value: event.Description})
	}
	event.Properties = setProperties(event.Properties, props)

	if _, err := io.WriteString(w, beginVEvent+crlf); err != nil {
		return err
	}

	if err := formatPropertiesList(w, event.Properties); err != nil {
		return err
	}

	for _, alarm := range event.Alarms {
		if err := formatAlarm(w, alarm); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, endVEvent+crlf)
	return err
}

func formatAlarm(w io.Writer, alarm *Alarm) error {
	var props []*Property
	if alarm.Action != "" {
		props = append(props, &Property{Name: "ACTION", Value: alarm.Action})
	}
	if alarm.Trigger != "" {
		props = append(props, &Property{Name: "TRIGGER", Value: alarm.Trigger})
	}
	alarm.Properties = setProperties(alarm.Properties, props)

	if _, err := io.WriteString(w, beginValarm+crlf); err != nil {
		return err
	}

	if err := formatPropertiesList(w, alarm.Properties); err != nil {
		return err
	}

	_, err := io.WriteString(w, endVAlarm+crlf)
	return err
}

func formatPropertiesList(w io.Writer, props []*Property) error {
	for _, prop := range props {
		if err := formatProperty(w, prop); err != nil {
			return err
		}
	}
	return nil
}

func formatProperty(w io.Writer, prop *Property) error {
	var buf bytes.Buffer
	buf.WriteString(prop.Name)

	for name, param := range prop.Params {
		buf.WriteString(";")
		buf.WriteString(name)
		buf.WriteString("=")
		for i, v := range param.Values {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(strconv.Quote(v))
		}
	}

	buf.WriteString(":")
	buf.WriteString(prop.Value)
	buf.WriteString(crlf)

	_, err := buf.WriteTo(w)
	return err
}

func setProperties(l []*Property, newProps []*Property) []*Property {
	m := make(map[string]*Property, len(newProps))
	for _, newProp := range newProps {
		m[newProp.Name] = newProp
	}

	for i, prop := range l {
		newProp, ok := m[prop.Name]
		if ok {
			l[i] = newProp
			delete(m, prop.Name)
		}
	}

	for _, newProp := range newProps {
		if _, ok := m[newProp.Name]; ok {
			l = append(l, newProp)
		}
	}

	return l
}
