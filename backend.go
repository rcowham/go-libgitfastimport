package libfastimport

import (
	"bufio"
	"io"

	"git.lukeshu.com/go/libfastimport/textproto"
)

// A Backend is something that consumes a fast-import stream; the
// Backend object provides methods for writing to it.
type Backend struct {
	w   *bufio.Writer
	fiw *textproto.FIWriter
	cbr *textproto.CatBlobReader

	onErr func(error) error
}

func NewBackend(fastImport io.Writer, catBlob io.Reader, onErr func(error) error) *Backend {
	ret := &Backend{}
	ret.w = bufio.NewWriter(fastImport)
	ret.fiw = textproto.NewFIWriter(ret.w)
	if catBlob != nil {
		ret.cbr = textproto.NewCatBlobReader(catBlob)
	}
	ret.onErr = onErr
	return ret
}

func (b *Backend) Do(cmd Cmd) error {
	err := cmd.fiWriteCmd(b.fiw)
	if err != nil {
		return b.onErr(err)
	}
	err = b.w.Flush()
	if err != nil {
		return b.onErr(err)
	}
	return nil
}

func (b *Backend) GetMark(cmd CmdGetMark) (sha1 string, err error) {
	err = b.Do(cmd)
	if err != nil {
		return
	}
	line, err := b.cbr.ReadLine()
	if err != nil {
		err = b.onErr(err)
		return
	}
	sha1, err = cbpGetMark(line)
	if err != nil {
		err = b.onErr(err)
	}
	return
}

func (b *Backend) CatBlob(cmd CmdCatBlob) (sha1 string, data string, err error) {
	err = b.Do(cmd)
	if err != nil {
		return
	}
	line, err := b.cbr.ReadLine()
	if err != nil {
		err = b.onErr(err)
		return
	}
	sha1, data, err = cbpCatBlob(line)
	if err != nil {
		err = b.onErr(err)
	}
	return
}

func (b *Backend) Ls(cmd CmdLs) (mode textproto.Mode, dataref string, path textproto.Path, err error) {
	err = b.Do(cmd)
	if err != nil {
		return
	}
	line, err := b.cbr.ReadLine()
	if err != nil {
		err = b.onErr(err)
		return
	}
	mode, dataref, path, err = cbpLs(line)
	if err != nil {
		err = b.onErr(err)
	}
	return
}
