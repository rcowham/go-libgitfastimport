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

package libfastimport

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func cbpGetMark(line string) (string, error) {
	if len(line) != 41 {
		return "", errors.Errorf("get-mark: short <sha1>\\n: %q", line)
	}
	if line[40] != '\n' {
		return "", errors.Errorf("get-mark: malformed <sha1>\\n: %q", line)
	}
	for _, b := range line[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return "", errors.Errorf("get-mark: malformed <sha1>: %q", line[:40])
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
		return "", "", errors.Errorf("cat-blob: missing trailing newline")
	}

	lf := strings.IndexByte(full, '\n')
	if lf < 0 || lf == len(full)-1 {
		return "", "", errors.Errorf("cat-blob: malformed header: %q", full)
	}
	head := full[:lf]
	data = full[lf+1 : len(full)-1]

	if len(head) < 40+6+1 {
		return "", "", errors.Errorf("cat-blob: malformed header: %q", head)
	}

	sha1 = head[:40]
	for _, b := range sha1 {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			return "", "", errors.Errorf("cat-blob: malformed <sha1>: %q", sha1)
		}
	}

	if string(head[40:46]) != " blob " {
		return "", "", errors.Errorf("cat-blob: malformed header: %q", head)
	}

	size, err := strconv.Atoi(head[46:])
	if err != nil {
		return "", "", errors.Wrap(err, "cat-blob: malformed blob size")
	}

	if size != len(data) {
		return "", "", errors.Errorf("cat-blob: size header (%d) didn't match delivered size (%d)", size, len(data))
	}

	return sha1, data, err
}

func cbpLs(line string) (mode Mode, dataref string, path Path, err error) {
	//     <mode> SP ('blob' | 'tree' | 'commit') SP <dataref> HT <path> LF
	// or
	//     'missing' SP <path> LF
	if line[len(line)-1] != '\n' {
		return 0, "", "", errors.New("ls: missing trailing newline")
	}
	line = line[:len(line)-1]

	if strings.HasPrefix(line, "missing ") {
		strPath := line[8:]
		return 0, "", PathUnescape(strPath), nil
	} else {
		fields := strings.SplitN(line, " ", 3)
		if len(fields) < 3 {
			return 0, "", "", errors.Errorf("ls: malformed line: %q", line)
		}
		ht := strings.IndexByte(fields[2], '\t')
		if ht < 0 {
			return 0, "", "", errors.Errorf("ls: malformed line: %q", line)
		}
		strMode := fields[0]
		//strType := fields[1]
		strRef := fields[2][:ht]
		strPath := fields[2][ht+1:]

		nMode, err := strconv.ParseUint(strMode, 8, 18)
		if err != nil {
			return 0, "", "", err
		}
		return Mode(nMode), strRef, PathUnescape(strPath), nil
	}
}
