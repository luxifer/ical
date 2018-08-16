# iCalendar lexer/parser

[![Build Status](https://travis-ci.org/luxifer/ical.svg?branch=master)](https://travis-ci.org/luxifer/ical)

Golang iCalendar lexer/parser implementing [RFC 5545](https://tools.ietf.org/html/rfc5545). This project is heavily inspired of the talk [Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE) by Rob Pike.

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
