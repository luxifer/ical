# iCalendar lexer/parser

[![Build Status](https://travis-ci.org/luxifer/ical.svg?branch=master)](https://travis-ci.org/luxifer/ical)

Golang iCalendar lexer/parser implementing [RFC 5545](https://tools.ietf.org/html/rfc5545).

## Usage

```go
import (
    "github.com/luxifer/ical"
)

// filename is an io.Reader
// second parameter is a *time.Location which defaults to system local
calendar, err := ical.Parse(filename, nil)
```

## TODO

* Implements Missing Properties on VEVENT
* Implements VTODO
* Implements VJOURNAL
* Implements VFREEBUSY
* Implements VTIMEZONE
* Implements Missing Components Properties
