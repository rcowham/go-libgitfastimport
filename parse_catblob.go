package libfastimport

import (
	"fmt"
	"strconv"
	"strings"

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
	data = full[lf+1 : len(full)-1]

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
	line = line[:len(line)-1]

	if strings.HasPrefix(line, "missing ") {
		strPath := line[8:]
		return 0, "", textproto.PathUnescape(strPath), nil
	} else {
		fields := strings.SplitN(line, " ", 3)
		if len(fields) < 3 {
			return 0, "", "", fmt.Errorf("ls: malformed line: %q", line)
		}
		ht := strings.IndexByte(fields[2], '\t')
		if ht < 0 {
			return 0, "", "", fmt.Errorf("ls: malformed line: %q", line)
		}
		strMode := fields[0]
		//strType := fields[1]
		strRef := fields[2][:ht]
		strPath := fields[2][ht+1:]

		nMode, err := strconv.ParseUint(strMode, 8, 18)
		if err != nil {
			return 0, "", "", err
		}
		return textproto.Mode(nMode), strRef, textproto.PathUnescape(strPath), nil
	}
}
