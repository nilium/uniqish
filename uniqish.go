// Copyright (c) 2015, Noel Cower.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// uniqish is a small utility for splitting input string(s) into a series of
// segments/lines and ensuring that only one of each segment appears. The first
// segment encountered, from right to left, is kept and repeats are removed.
// This is similar to uniq, except that it doesn't only consider adjacent
// lines.
//
//     $ go get go.spiff.io/uniqish
//
// By default, if no strings are passed on the command line, uniqish will
// attempt to read strings from standard input.
//
// A simple example case for using uniqish is when sourcing a shell profile,
// but wanting to remove re-occurring entries:
//
//     PATH="$(uniqish -s ':' "$PATH")"
//
// This can be used to clean up a PATH with duplicate entries. I do this mainly
// because I keep a small dev environment on an SD card and regularly source
// a PROFILE file on it that sets up everything for that shell session. It's
// not likely to be useful in a lot of situations.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var exitStatus int = 0
var delim byte
var delimStr string
var delimBuffer [1]byte
var delimSlice = delimBuffer[:]
var found = map[string]bool{}

var out = bufio.NewWriter(os.Stdout)

// isTTY attempts to determine whether the current stdout refers to a terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting Stat of os.Stdout:", err)
		return true // Assume human readable
	}
	return (fi.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe
}

func makeUniqish(r io.Reader, w io.Writer) {
	br := bufio.NewReader(r)
	eof := false

	for !eof {
		var line string
		line, err := br.ReadString(delim)
		eof = err == io.EOF
		if err != nil && !eof {
			log.Println("error reading input segment:", err)
			break
		} else if len(line) == 0 {
			continue
		}

		// Strip delimiter
		if line[len(line)-1] == delim {
			line = line[:len(line)-1]
			if len(line) == 0 {
				continue
			}
		}

		// Filter existing lines
		if found[line] {
			continue
		}

		if len(found) > 0 {
			if _, err := w.Write(delimSlice); err != nil {
				log.Panic(err)
			}
		}

		if _, err := io.WriteString(w, line); err != nil {
			log.Panic(err)
		}

		found[line] = true
	}
}

func main() {
	defer func() { os.Exit(exitStatus) }()
	defer out.Flush()

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "uniqish [OPTIONS] [STRINGS...]\n\nOPTIONS:")
		flag.PrintDefaults()
	}
	flag.StringVar(&delimStr, "s", "\n", "the string separator (defaults to newline)")
	flag.Parse()
	if len(delimStr) != 1 {
		log.Fatalf("separator must be one byte in length (sep=%q)", delimStr)
		return
	}

	delim = delimStr[0]
	delimBuffer[0] = delim

	if isTTY() {
		defer out.WriteByte('\n')
	}

	if flag.NArg() == 0 {
		makeUniqish(os.Stdin, out)
	} else {
		for _, str := range flag.Args() {
			strbuf := strings.NewReader(str)
			makeUniqish(strbuf, out)
		}
	}
}
