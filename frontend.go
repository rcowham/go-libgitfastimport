package libfastimport

import (
	"io"
	"bufio"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type Frontend struct {
	w *bufio.Writer
	fiw *textproto.FIWriter
	cbr *textproto.CatBlobReader
}

func NewFrontend(w io.Writer, r io.Reader) *Frontend {
	ret := Frontend{}
	ret.w = bufio.NewWriter(w)
	ret.fiw = textproto.NewFIWriter(ret.w)
	if r != nil {
		ret.cbr = textproto.NewCatBlobReader(r)
	}
	return &ret
}

func (f *Frontend) Do(cmd Cmd) error {
	err := cmd.fiWriteCmd(f.fiw)
	if err != nil {
		return err
	}
	return f.w.Flush()
}

func (f *Frontend) GetMark(cmd CmdGetMark) (string, error) {
	err := f.Do(cmd)
	if err != nil {
		return "", err
	}
	line, err := f.cbr.ReadLine()
	if err != nil {
		return "", err
	}
	return cbpGetMark(line)
}

func (f *Frontend) CatBlob(cmd CmdCatBlob) (sha1 string, data string, err error) {
	err = f.Do(cmd)
	if err != nil {
		return "", "", err
	}
	line, err := f.cbr.ReadLine()
	if err != nil {
		return "", "", err
	}
	return cbpCatBlob(line)
}

func (f *Frontend) Ls(cmd CmdLs) (mode textproto.Mode, dataref string, path textproto.Path, err error) {
	err = f.Do(cmd)
	if err != nil {
		return 0, "", "", err
	}
	line, err := f.cbr.ReadLine()
	if err != nil {
		return 0, "", "", err
	}
	return cbpLs(line)
}
