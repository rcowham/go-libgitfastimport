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
}

func NewFIReader(r io.Reader) *FIReader {
	return &FIReader{
		r: bufio.NewReader(r),
	}
}

func (fir *FIReader) ReadLine() (line string, err error) {
retry:
	line, err = fir.r.ReadString('\n')
	if err != nil {
		return
	}
	if len(line) == 1 {
		goto retry
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

func (fiw *FIWriter) WriteData(data []byte) error {
	err := fiw.WriteLine("data", len(data))
	if err != nil {
		return err
	}
	_, err = fiw.w.Write(data)
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
retry:
	line, err = cbr.r.ReadString('\n')
	if err != nil {
		return
	}
	if len(line) == 1 {
		goto retry
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
	return &CatBlobWriter {
		w: w,
	}
}

func (cbw *CatBlobWriter) WriteLine(a ...interface{}) error {
	_, err := fmt.Fprintln(cbw.w, a...)
	return err
}

func (cbw *CatBlobWriter) WriteBlob(sha1 string, data []byte) error {
	err := cbw.WriteLine(sha1, "blob", len(data))
	if err != nil {
		return err
	}
	_, err = cbw.w.Write(data)
	return err
}
