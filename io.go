package libfastimport

import (
	"fmt"
	"bufio"
	"bytes"
	"io"
	"strconv"
)

type FIReader struct {
	r *bufio.Reader
}

func NewFIReader(r io.Reader) *FIReader {
	return &FIReader{
		r: bufio.NewReader(r),
	}
}

func (fir *FIReader) ReadSlice() (line []byte, err error) {
retry:
	line, err = fir.r.ReadSlice('\n')
	if err != nil {
		return
	}
	if len(line) == 1 {
		goto retry
	}

	if bytes.HasPrefix(line, []byte("data ")) {
		if string(line[5:7]) == "<<" {
			// Delimited format
			delim := line[7 : len(line)-1]
			suffix := []byte("\n" + string(delim) + "\n")

			_line := make([]byte, len(line))
			copy(_line, line)
			line = _line

			for !bytes.HasSuffix(line, suffix) {
				_line, err = fir.r.ReadSlice('\n')
				line = append(line, _line...)
				if err != nil {
					return
				}
			}
		} else {
			// Exact byte count format
			var size int
			size, err = strconv.Atoi(string(line[5 : len(line)-1]))
			if err != nil {
				return
			}
			_line := make([]byte, size+len(line))
			copy(_line, line)
			var n int
			n, err = io.ReadFull(fir.r, _line[len(line):])
			line = _line[:n+len(line)]
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

func (cbr *CatBlobReader) ReadSlice() (line []byte, err error) {
retry:
	line, err = cbr.r.ReadSlice('\n')
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
	if len(line) <= 46 || string(line[40:46]) != " blob " {
		return
	}
	for _, b := range line[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return
		}
	}
	// we have a cat-blob result
	var size int
	size, err = strconv.Atoi(string(line[46 : len(line)-1]))
	if err != nil {
		return
	}
	_line := make([]byte, len(line)+size+1)
	copy(_line, line)
	n, err := io.ReadFull(cbr.r, _line[len(line):])
	line = _line[:n+len(line)]
	return
}

type CatBlobWriter struct {
	w io.Writer
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
