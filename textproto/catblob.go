package textproto

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

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
	// cat-blob : <sha1> SP 'blob' SP <size> LF
	//            <data> LF
	// ls       : <mode> SP ('blob' | 'tree' | 'commit') SP <dataref> HT <path> LF
	// ls       : 'missing' SP <path> LF

	// decide if we have a cat-blob result (return early if we don't)
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
