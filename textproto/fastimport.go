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
