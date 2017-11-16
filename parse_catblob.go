package libfastimport

import (
	"fmt"
	"strings"
	"strconv"

	"git.lukeshu.com/go/libfastimport/textproto"
)

func cbpGetMark(line string) (string, error) {
	if len(line) != 41 {
		return "", fmt.Errorf("get-mark: short <sha1>\\n: %q", line)
	}
	if line[40] != '\n' {
		return "", fmt.Errorf("get-mark: malformed <sha1>\\n: %q", line)
	}
	for _, b := range line[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return "", fmt.Errorf("get-mark: malformed <sha1>: %q", line[:40])
		}
	}
	return line[:40], nil
}

func cbpCatBlob(full string) (sha1 string, data string, err error) {
	// The format is:
	//
	//    <sha1> SP 'blob' SP <size> LF
	//    <data> LF

	if full[len(full)-1] != '\n' {
		return "", "", fmt.Errorf("cat-blob: missing trailing newline")
	}

	lf := strings.IndexByte(full, '\n')
	if lf < 0 || lf == len(full)-1 {
		return "", "", fmt.Errorf("cat-blob: malformed header: %q", full)
	}
	head := full[:lf]
	data = full[lf+1:len(full)-1]

	if len(head) < 40+6+1 {
		return "", "", fmt.Errorf("cat-blob: malformed header: %q", head)
	}

	sha1 = head[:40]
	for _, b := range sha1 {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return "", "", fmt.Errorf("cat-blob: malformed <sha1>: %q", sha1)
		}
	}

	if string(head[40:46]) != " blob " {
		return "", "", fmt.Errorf("cat-blob: malformed header: %q", head)
	}

	size, err := strconv.Atoi(head[46:])
	if err != nil {
		return "", "", fmt.Errorf("cat-blob: malformed blob size: %v", err)
	}

	if size != len(data) {
		return "", "", fmt.Errorf("cat-blob: size header (%d) didn't match delivered size (%d)", size, len(data))
	}

	return sha1, data, err
}

func cbpLs(line string) (mode textproto.Mode, dataref string, path textproto.Path, err error) {
	//     <mode> SP ('blob' | 'tree' | 'commit') SP <dataref> HT <path> LF
	// or
	//     'missing' SP <path> LF
	if line[len(line)-1] != '\n' {
		return 0, "", "", fmt.Errorf("ls: missing trailing newline")
	}
	if strings.HasPrefix(line, "missing ") {
		strPath := line[8:len(line)-1]
		return 0, "", textproto.PathUnescape(strPath), nil
	} else {
		sp1 := strings.IndexByte(line, ' ')
		sp2 := strings.IndexByte(line[sp1+1:], ' ')
		ht := strings.IndexByte(line[sp2+1:], '\t')
		if sp1 < 0 || sp2 < 0 || ht < 0 {
			return 0, "", "", fmt.Errorf("ls: malformed line: %q", line)
		}
		strMode := line[:sp1]
		strRef := line[sp2+1:ht]
		strPath := line[ht+1:len(line)-1]

		nMode, err := strconv.ParseUint(strMode, 8, 18)
		if err != nil {
			return 0, "", "", err
		}
		return textproto.Mode(nMode), strRef, textproto.PathUnescape(strPath), nil
	}
}
