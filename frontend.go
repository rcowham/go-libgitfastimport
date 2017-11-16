package libfastimport

import (
	"bufio"
	"io"
	"strings"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type UnsupportedCommand string

func (e UnsupportedCommand) Error() string {
	return "Unsupported command: " + string(e)
}

type cmderror struct {
	Cmd
	error
}

// A Frontend is something that produces a fast-import stream; the
// Frontend object provides methods for reading from it.
type Frontend struct {
	fir *textproto.FIReader
	cbw *textproto.CatBlobWriter
	w   *bufio.Writer
	c   chan cmderror
}

func NewFrontend(fastImport io.Reader, catBlob io.Writer) *Frontend {
	ret := Frontend{}
	ret.fir = textproto.NewFIReader(fastImport)
	if catBlob != nil {
		ret.w = bufio.NewWriter(catBlob)
		ret.cbw = textproto.NewCatBlobWriter(ret.w)
	}
	return &ret
}

func (f *Frontend) nextLine() (line string, err error) {
retry:
	line, err = f.fir.ReadLine()
	if err != nil {
		return
	}
	switch {
	case strings.HasPrefix(line, "#"):
		f.c <- cmderror{CmdComment{Comment: line[1:]}, nil}
		goto retry
	case strings.HasPrefix(line, "cat-blob "):
		f.c <- parse_cat_blob(line)
		goto retry
	case strings.HasPrefix(line, "get-mark "):
		f.c <- parse_get_mark(line)
		goto retry
	default:
		return
	}
}

func (f *Frontend) parse() {
	for {
		line, err := f.nextLine()
		if err != nil {
			f.c <- cmderror{nil, err}
			return
		}
		switch {
		case strings.HasPrefix(line, "blob "):
		case strings.HasPrefix(line, "commit "):
		case strings.HasPrefix(line, "checkpoint\n"):
		case strings.HasPrefix(line, "done\n"):
		case strings.HasPrefix(line, "feature "):
		case strings.HasPrefix(line, "ls "):
		case strings.HasPrefix(line, "option "):
		case strings.HasPrefix(line, "progress "):
		case strings.HasPrefix(line, "reset "):
		case strings.HasPrefix(line, "tag "):
		default:
			f.c <- cmderror{nil, UnsupportedCommand(line)}
			return
		}
	}
}

func (f *Frontend) ReadCmd() (Cmd, error) {
	cmderror := <-f.c
	cmd := cmderror.Cmd
	err := cmderror.error
	return cmderror.Cmd, cmderror.error
}

func (f *Frontend) RespondGetMark(sha1 string) error {
	// TODO
	return f.w.Flush()
}

func (f *Frontend) RespondCatBlob(sha1 string, data string) error {
	// TODO
	return f.w.Flush()
}

func (f *Frontend) RespondLs(mode textproto.Mode, dataref string, path textproto.Path) error {
	// TODO
	return f.w.Flush()
}
