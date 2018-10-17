package ical

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

var calendarList = []string{"fixtures/example.ics", "fixtures/with-alarm.ics", "fixtures/facebookbirthday.ics"}

func TestParse(t *testing.T) {
	for _, filename := range calendarList {
		file, _ := os.Open(filename)
		_, err := Parse(file, nil)
		file.Close()

		if err != nil {
			t.Error(fmt.Errorf("%v on '%s'", err, filename))
		}
	}
}

func Test_parseDate(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	type args struct {
		prop *Property
		l    *time.Location
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Property with only date layout",
			args: args{
				prop: &Property{
					Name: "DTSTART",
					Params: map[string]*Param{
						"VALUE": &Param{
							Values: []string{"DATE"},
						},
					},
					Value: "19980119",
				},
				l: time.Local,
			},
			want: time.Date(1998, time.January, 19, 0, 0, 0, 0, time.Local),
		},
		{
			name: "Floating (local) date-time (no time zone)",
			args: args{
				prop: &Property{
					Name: "DTSTART",
					Params: map[string]*Param{
						"TZID": &Param{
							Values: []string{"America/New_York"},
						},
					},
					Value: "19980119T020000",
				},
				l: time.Local,
			},
			want: time.Date(1998, time.January, 19, 2, 0, 0, 0, loc),
		},
		{
			name: "Date with no layout indication",
			args: args{
				prop: &Property{
					Name:  "DSTART",
					Value: "19980119T020000",
				},
				l: time.Local,
			},
			want: time.Date(1998, time.January, 19, 2, 0, 0, 0, time.Local),
		},
		{
			name: "Date with no layout and in UTC",
			args: args{
				prop: &Property{
					Name:  "DSTART",
					Value: "19980119T070000Z",
				},
				l: time.Local,
			},
			want: time.Date(1998, time.January, 19, 7, 0, 0, 0, time.UTC),
		},
		{
			name: "Property with date-time layout",
			args: args{
				prop: &Property{
					Name: "DSTART",
					Params: map[string]*Param{
						"VALUE": &Param{
							Values: []string{"DATE-TIME"},
						},
					},
					Value: "19980119T070000",
				},
				l: time.Local,
			},
			want: time.Date(1998, time.January, 19, 7, 0, 0, 0, time.Local),
		},
		{
			name: "Datetime with bad timezone",
			args: args{
				prop: &Property{
					Name: "DTSTART",
					Params: map[string]*Param{
						"TZID": &Param{
							Values: []string{"Z"},
						},
					},
					Value: "19980119T020000",
				},
				l: time.Local,
			},
			want: time.Date(1998, time.January, 19, 2, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.args.prop, tt.args.l)
			if (err != nil) != false {
				t.Errorf("parseDate() error = %v, wantErr %v", err, false)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
