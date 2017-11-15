package libfastimport

import (
	"fmt"
	"bytes"
	"strconv"
)

func CatBlobParseGetMark(dat []byte) (string, error) {
	if len(dat) != 41 {
		return "", fmt.Errorf("get-mark: short <sha1>\\n: %q", string(dat))
	}
	if dat[40] != '\n' {
		return "", fmt.Errorf("get-mark: malformed <sha1>\\n: %q", string(dat))
	}
	for _, b := range dat[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return "", fmt.Errorf("get-mark: malformed <sha1>: %q", string(dat[:40]))
		}
	}
	return string(dat[:40]), nil
}

func CatBlobParseCatBlob(full []byte) (sha1 string, data []byte, err error) {
	// The format is:
	//
	//    <sha1> SP 'blob' SP <size> LF
	//    <data> LF

	lf := bytes.IndexByte(full, '\n')
	if lf < 0 || lf == len(full)-1 {
		return "", nil, fmt.Errorf("cat-blob: malformed header: %q", string(full))
	}
	head := full[:lf]
	data = full[lf+1:len(full)-1]

	if len(head) < 40+6+1 {
		return "", nil, fmt.Errorf("cat-blob: malformed header: %q", string(head))
	}

	sha1 = string(head[:40])
	for _, b := range sha1 {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return "", nil, fmt.Errorf("cat-blob: malformed <sha1>: %q", sha1)
		}
	}

	if string(head[40:46]) != " blob " {
		return "", nil, fmt.Errorf("cat-blob: malformed header: %q", head)
	}

	size, err := strconv.Atoi(string(head[46:]))
	if err != nil {
		return "", nil, fmt.Errorf("cat-blob: malformed blob size: %v", err)
	}

	if size != len(data) {
		return "", nil, fmt.Errorf("cat-blob: size header (%d) didn't match delivered size (%d)", size, len(data))
	}

	return sha1, data, err
}

func CatBlobParseLs(dataref string, path string) error {
	//     <mode> SP ('blob' | 'tree' | 'commit') SP <dataref> HT <path> LF
	// or
	//     'missing' SP <path> LF
	return nil // TODO
}
