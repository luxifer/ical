# iCalendar lexer/parser

[![Build Status](https://travis-ci.org/Xotelia/ical.svg?branch=master)](https://travis-ci.org/Xotelia/ical)

Golang iCalendar lexer/parser based on RFC5545.

## Usage

```go
import (
    "github.com/Xotelia/ical"
)

// filename is an io.Reader
calendar, err := ical.Parse(filename)
```

## TODO

* Implements Missing Properties on VEVENT
* Implements VTODO
* Implements VJOURNAL
* Implements VFREEBUSY
* Implements VTIMEZONE
* Implements VALARM
* Implements Missing Components Properties
