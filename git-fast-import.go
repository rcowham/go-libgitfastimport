package libfastimport

import (
	"time"
)

type FastImport struct {
	w   *FIWriter
	r   *CatBlobReader
	err error
}

func (fi *FastImport) GetMark(mark int) (string, error) {
	if fi.printf("get-mark :%d\n", mark) != nil {
		return "", fi.ioErr
	}
	var dat [41]byte
	var n int
	_, fi.ioErr = io.ReadFull(fi.r, dat[:])
	if fi.ioErr != nil {
		return "", fi.ioErr
	}
	if dat[40] != '\n' {
		fi.ioErr = fmt.Errorf("get-mark: malformed <sha1>\\n: %q", string(dat))
		return "", fi.ioErr
	}
	for _, b := range dat[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			fi.ioErr = fmt.Errorf("get-mark: malformed <sha1>: %q", string(dat[:40]))
			return "", fi.ioErr
		}
	}
	return string(dat[:40])
}

func (fi *FastImport) CatBlob(dataref string) (sha1 string, data []byte, err error) {
	if fi.println("cat-blob %s\n", dataref) != nil {
		return "", nil, fi.ioErr
	}
	//    <sha1> SP 'blob' SP <size> LF
	//
	// That comes out to be 47+len(itoa(size)).  Assuming that
	// size is at most a 64-bit integer (a safe assumption), then
	// its limit is 20 digits.
	var head [67]byte
	i := 0
	for {
		_, fi.ioErr = io.ReadFull(fi.r, head[i:i+1])
		if fi.ioErr != nil {
			return "", nil, fi.ioErr
		}
		if head[i] == '\n' {
			break
		}
		i++
		if i == len(head) {
			fi.ioErr = fmt.Errorf("cat-blob: overly-long header line: %q", string(head))
			return "", nil, fi.ioErr
		}
	}
	i--
	if head[i] != '\n' {
		panic("wut")
	}
	for _, b := range head[:40] {
		if !(('0' <= b && b <= '9') || ('a' <= b && b <= 'f')) {
			fi.ioErr = fmt.Errorf("cat-blob: malformed <sha1>: %q", string(head[:40]))
			return "", nil, fi.ioErr
		}
	}
	if string(head[40:46]) != " blob " {
		fi.ioErr = fmt.Errorf("cat-blob: malformed header: %q", string(head[:i]))
		return "", nil, fi.ioErr
	}
	size, err := strconv.Atoi(string(head[46:i]))
	if err != nil {
		fi.ioErr = fmt.Errorf("cat-blob: malformed blob size: %v", err)
		return "", nil, fi.ioErr
	}
	dat := make([]byte, size+1)
	_, fi.ioErr = io.ReadFull(fi.r, dat)
	if dat[size] != '\n' {
		fi.ioErr = fmt.Errorf("cat-blob: expected newline after data")
		return "", nil, fi.ioErr
	}
	return string(head[:40]), dat[:size], nil
}

func (fi *FastImport) Ls(dataref string, path string) error {
	if dataref == "" {
		fi.printf("ls %s\n", quotePath(path))
	} else {
		fi.printf("ls %s %s\n", dataref, quotePath(path))
	}
	if fi.ioErr != nil {
		return fi.ioErr
	}
	k
}
func (fi *FastImport) Feature() error
func (fi *FastImport) Option() error

func (fi *FastImport) Done() error {
	fi.printf("done\n")
	if fi.ioErr == nil {
		fi.ioErr = w.Close()
	}
	return fi.ioErr
}

func init() {
	x := exec.Cmd{
		Path: prog["git"],
		Args: {"git", "fast-import",
			"--done",
			"--cat-blob-fd=" + strconv.Itoa(TODO)},
	}
}
