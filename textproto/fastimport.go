// Copyright (C) 2017-2018  Luke Shumaker <lukeshu@lukeshu.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package textproto

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FIReader is a low-level parser of a fast-import stream.
type FIReader struct {
	r *bufio.Reader

	line *string
	err  error
}

// NewFIReader creates a new FIReader parser.
func NewFIReader(r io.Reader) *FIReader {
	return &FIReader{
		r: bufio.NewReader(r),
	}
}

// ReadLine reads a "line" from the stream; with special handling for
// the "data" command, which isn't really a single line, but rather
// contains arbitrary binary data.
func (fir *FIReader) ReadLine() (line string, err error) {
	for len(line) <= 1 {
		line, err = fir.r.ReadString('\n')
		if err != nil {
			return
		}
	}

	if strings.HasPrefix(line, "data ") {
		if line[5:7] == "<<" {
			// Delimited format
			delim := line[7 : len(line)-1]
			suffix := "\n" + delim + "\n"

			for !strings.HasSuffix(line, suffix) {
				var _line string
				_line, err = fir.r.ReadString('\n')
				line += _line
				if err != nil {
					return
				}
			}
		} else {
			// Exact byte count format
			var size int
			size, err = strconv.Atoi(line[5 : len(line)-1])
			if err != nil {
				return
			}
			data := make([]byte, size)
			_, err = io.ReadFull(fir.r, data)
			line += string(data)
		}
	}
	return
}

// FIWriter is a low-level marshaller of a fast-import stream.
type FIWriter struct {
	w io.Writer
}

// NewFIWriter creates a new FIWriter marshaller.
func NewFIWriter(w io.Writer) *FIWriter {
	return &FIWriter{
		w: w,
	}
}

// WriteLine writes an ordinary line to the stream; arguments are
// handled similarly to fmt.Println.
func (fiw *FIWriter) WriteLine(a ...interface{}) error {
	_, err := fmt.Fprintln(fiw.w, a...)
	return err
}

// WriteData writes a 'data' command to the stream.
func (fiw *FIWriter) WriteData(data string) error {
	err := fiw.WriteLine("data", len(data))
	if err != nil {
		return err
	}
	_, err = io.WriteString(fiw.w, data)
	if err == nil && data[len(data)-1] != '\n' {
		_, err = io.WriteString(fiw.w, "\n")
	}
	return err
}
