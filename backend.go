package libfastimport

import (
	"io"
	"bufio"

	"git.lukeshu.com/go/libfastimport/textproto"
)

// A Backend is something that consumes a fast-import stream; the
// Backend object provides methods for writing to it.
type Backend struct {
	w *bufio.Writer
	fiw *textproto.FIWriter
	cbr *textproto.CatBlobReader
}

func NewBackend(fastImport io.Writer, catBlob io.Reader) *Backend {
	ret := Backend{}
	ret.w = bufio.NewWriter(fastImport)
	ret.fiw = textproto.NewFIWriter(ret.w)
	if catBlob != nil {
		ret.cbr = textproto.NewCatBlobReader(catBlob)
	}
	return &ret
}

func (b *Backend) Do(cmd Cmd) error {
	err := cmd.fiWriteCmd(b.fiw)
	if err != nil {
		return err
	}
	return b.w.Flush()
}

func (b *Backend) GetMark(cmd CmdGetMark) (string, error) {
	err := b.Do(cmd)
	if err != nil {
		return "", err
	}
	line, err := b.cbr.ReadLine()
	if err != nil {
		return "", err
	}
	return cbpGetMark(line)
}

func (b *Backend) CatBlob(cmd CmdCatBlob) (sha1 string, data string, err error) {
	err = b.Do(cmd)
	if err != nil {
		return "", "", err
	}
	line, err := b.cbr.ReadLine()
	if err != nil {
		return "", "", err
	}
	return cbpCatBlob(line)
}

func (b *Backend) Ls(cmd CmdLs) (mode textproto.Mode, dataref string, path textproto.Path, err error) {
	err = b.Do(cmd)
	if err != nil {
		return 0, "", "", err
	}
	line, err := b.cbr.ReadLine()
	if err != nil {
		return 0, "", "", err
	}
	return cbpLs(line)
}
