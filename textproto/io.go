// Copyright (C) 2017  Luke Shumaker <lukeshu@lukeshu.com>
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

type FIReader struct {
	r *bufio.Reader

	line *string
	err  error
}

func NewFIReader(r io.Reader) *FIReader {
	return &FIReader{
		r: bufio.NewReader(r),
	}
}

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

type FIWriter struct {
	w io.Writer
}

func NewFIWriter(w io.Writer) *FIWriter {
	return &FIWriter{
		w: w,
	}
}

func (fiw *FIWriter) WriteLine(a ...interface{}) error {
	_, err := fmt.Fprintln(fiw.w, a...)
	return err
}

func (fiw *FIWriter) WriteData(data string) error {
	err := fiw.WriteLine("data", len(data))
	if err != nil {
		return err
	}
	_, err = io.WriteString(fiw.w, data)
	return err
}

type CatBlobReader struct {
	r *bufio.Reader
}

func NewCatBlobReader(r io.Reader) *CatBlobReader {
	return &CatBlobReader{
		r: bufio.NewReader(r),
	}
}

func (cbr *CatBlobReader) ReadLine() (line string, err error) {
	for len(line) <= 1 {
		line, err = cbr.r.ReadString('\n')
		if err != nil {
			return
		}
	}

	// get-mark : <sha1> LF
	// cat-blob : <sha1> SP 'blob' SP <size> LF <data> LF
	// ls       : <mode> SP ('blob' | 'tree' | 'commit') SP <dataref> HT <path> LF
	// ls       : 'missing' SP <path> LF

	// decide if we have a cat-blob result
	if len(line) <= 46 || line[40:46] != " blob " {
		return
	}
	for _, b := range line[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return
		}
	}
	// we have a cat-blob result
	var size int
	size, err = strconv.Atoi(line[46 : len(line)-1])
	if err != nil {
		return
	}
	data := make([]byte, size+1)
	_, err = io.ReadFull(cbr.r, data)
	line += string(data[:size])
	return
}

type CatBlobWriter struct {
	w io.Writer
}

func NewCatBlobWriter(w io.Writer) *CatBlobWriter {
	return &CatBlobWriter{
		w: w,
	}
}

func (cbw *CatBlobWriter) WriteLine(a ...interface{}) error {
	_, err := fmt.Fprintln(cbw.w, a...)
	return err
}

func (cbw *CatBlobWriter) WriteBlob(sha1 string, data string) error {
	err := cbw.WriteLine(sha1, "blob", len(data))
	if err != nil {
		return err
	}
	_, err = io.WriteString(cbw.w, data)
	return err
}
