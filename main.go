// Copyright 2014 The timestamp authors
// SPDX-License-Identifier: BSD-3-Clause

// The timestamp tool prints timestamps in various formats.
package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"

	"willnorris.com/go/newbase60"
)

const day = 24 * time.Hour
const year = 365 * day

// simple, readable time format.  Based off of Time.String() without subseconds
const stdFormat = "2006-01-02 15:04:05 -0700 MST"

var (
	epoch = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	// inputFormats identifies the time formats used to parse user input.
	inputFormats = []string{
		time.RFC3339,
		"2006-01-02",
		"2006-002",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	//flags
	utc            = flag.Bool("utc", false, "parse times without timezones as UTC")
	printRFC3339   = flag.Bool("rfc3339", false, "print rfc 3339 timestamp only")
	printEpochDays = flag.Bool("epoch", false, "print sexigesimal epoch days only")
	printMilli     = flag.Bool("milli", false, "print unix timestamp in milliseconds")
	printNano      = flag.Bool("nano", false, "print unix timestamp in nanoseconds")
)

func usage() {
	const text = `timestamp is a tool for printing timestamps in various formats.

Usage:
  timestamp [-utc] [-rfc3339] [-epoch] [time]

timestamp will print the specified time in the following formats:
  - unix timestamp (number of seconds since January 1, 1970 UTC)
  - rfc 3339 timestamp in the specified timezone (if not UTC)
  - rfc 3339 timestamp in local timezone (if not specified tz or UTC)
  - rfc 3339 timestamp in UTC
  - ordinal date (year and day of the year) in the specified timezone (if not UTC)
  - ordinal date (year and day of the year) in UTC
  - epoch days (number of days since January 1, 1970 UTC) as decimal and
    sexigesimal (newbase60) formatted. This is only printed if date is after
    1970-01-01, and is always calculated based on UTC time.

time can be specified as a full rfc 3339 timestamp, just the date component
(YYYY-MM-DD), an ordinal date (YYYY-DDD), or as newbase60 encoded epoch days.
If no time is specified, the current system time will be used.

time values without an explicit timezone will be interpreted as the local
system timezone unless the -utc flag is provided.

Flags:
`

	fmt.Fprint(os.Stderr, text)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	loc := time.Local
	if *utc {
		loc = time.UTC
	}

	t, err := parseInput(flag.Arg(0), loc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
	}

	printOutput(os.Stdout, t, loc)
}

func parseInput(s string, loc *time.Location) (t time.Time, err error) {
	if s == "" {
		return time.Now().In(loc), nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		d := int64(f)
		var n int64

		// If greater than 10 digits, we're either dealing with a unix timestamp after 2286-11-20
		// or we have a timestamp after 2001-09-08 with sub-second precision.
		// We assume the latter and treat anything after the 10th digit as fractional seconds.
		if len(s) > 10 {
			if len(s) > 19 {
				// 19 digits is nanosecond precision, which is all Go supports
				s = s[:19]
			}
			d, _ = strconv.ParseInt(s[:10], 10, 64)
			n, _ = strconv.ParseInt(s[10:], 10, 64)
			n *= int64(math.Pow10(19 - len(s)))
		}
		return time.Unix(d, n).In(loc), nil
	}

	for _, f := range inputFormats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t, nil
		}
	}

	i := newbase60.DecodeToInt(s)
	t = epoch.Add(time.Duration(i) * day)
	if t.Year() < 1970 || time.Now().Add(100*year).Year() < t.Year() {
		err = fmt.Errorf("Parsed %q as a newbase60 epoch date outside of normal bounds. This might be an error.", s)
	}
	return t, err
}

func printOutput(w io.Writer, t time.Time, loc *time.Location) {
	epochDays := int(t.UTC().Sub(epoch) / day)

	if *printRFC3339 {
		fmt.Fprint(w, t.Format(time.RFC3339))
		return
	}

	if *printEpochDays {
		fmt.Fprint(w, newbase60.EncodeInt(epochDays))
		return
	}

	if *printMilli {
		fmt.Fprint(w, t.UnixNano()/1e3)
		return
	}

	if *printNano {
		fmt.Fprint(w, t.UnixNano())
		return
	}

	fmt.Fprintf(w, "%s\n\n", t.Format(stdFormat))
	printTime(w, "Unix Timestamp", "%d", t.Unix())

	if t.Location() != time.UTC {
		printTime(w, "RFC 3339", "%s", t.Format(time.RFC3339))
	}
	if t.Location() != time.Local {
		printTime(w, "RFC 3339 (Local)", "%s", t.Local().Format(time.RFC3339))
	}
	printTime(w, "RFC 3339 (UTC)", "%s", t.UTC().Format(time.RFC3339))

	if t.Location() != time.UTC {
		printTime(w, "Ordinal Date", "%d-%03d", t.Year(), t.YearDay())
	}
	printTime(w, "Ordinal Date (UTC)", "%d-%03d", t.UTC().Year(), t.UTC().YearDay())

	if epochDays > 0 {
		printTime(w, "Epoch Days", "%d (%s)", epochDays, newbase60.EncodeInt(epochDays))
	}
}

func printTime(w io.Writer, name, format string, a ...interface{}) {
	fmt.Fprintf(w, "%-19s ", name+":")
	fmt.Fprintf(w, format+"\n", a...)
}
