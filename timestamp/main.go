// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// The timestamp tool prints timestamps in various formats.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"willnorris.com/go/newbase60"
)

const day = 24 * time.Hour

var (
	epoch   = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	formats = []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	//flags
	utc = flag.Bool("utc", false, "parse input as UTC")
)

func main() {
	flag.Parse()

	loc := time.Local
	if *utc {
		loc = time.UTC
	}

	t := parseInput(flag.Arg(0), loc)

	if t.IsZero() {
		fmt.Fprintln(os.Stderr, "Unable to parse timestamp")
		os.Exit(1)
	}

	fmt.Printf("%s\n\n", t)
	fmt.Printf("Unix Timestamp: %d\n", t.Unix())
	if t.Location() != time.UTC {
		fmt.Printf("RFC 3339:       %s\n", t.Format(time.RFC3339))
	}
	fmt.Printf("RFC 3339 (UTC): %s\n", t.UTC().Format(time.RFC3339))
	fmt.Printf("Ordinal Date:   %d-%d\n", t.Year(), t.YearDay())

	epochDays := int(t.UTC().Sub(epoch) / day)
	fmt.Printf("Epoch Days:     %d (%s)\n", epochDays, newbase60.EncodeInt(epochDays))
}

func parseInput(s string, loc *time.Location) time.Time {
	if s == "" {
		return time.Now()
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(i, 0)
	}

	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t
		}
	}

	return time.Time{}
}
