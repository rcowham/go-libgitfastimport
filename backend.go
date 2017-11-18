package libfastimport

import (
	"bufio"
	"io"

	"git.lukeshu.com/go/libfastimport/textproto"
)

// A Backend is something that consumes a fast-import stream; the
// Backend object provides methods for writing to it.
type Backend struct {
	w   io.WriteCloser
	bw  *bufio.Writer
	fiw *textproto.FIWriter
	cbr *textproto.CatBlobReader

	err   error
	onErr func(error) error
}

func NewBackend(fastImport io.WriteCloser, catBlob io.Reader, onErr func(error) error) *Backend {
	ret := &Backend{}
	ret.w = fastImport
	ret.bw = bufio.NewWriter(ret.w)
	ret.fiw = textproto.NewFIWriter(ret.bw)
	if catBlob != nil {
		ret.cbr = textproto.NewCatBlobReader(catBlob)
	}
	ret.onErr = func(err error) error {
		ret.err = err

		// Close the underlying writer, but don't let the
		// error mask the previous error.
		err = ret.w.Close()
		if ret.err == nil {
			ret.err = err
		}

		if onErr != nil {
			ret.err = onErr(ret.err)
		}
		return ret.err
	}
	return ret
}

func (b *Backend) Do(cmd Cmd) error {
	if b.err == nil {
		return b.err
	}

	err := cmd.fiWriteCmd(b.fiw)
	if err != nil {
		return b.onErr(err)
	}
	err = b.bw.Flush()
	if err != nil {
		return b.onErr(err)
	}

	if _, ok := cmd.(CmdDone); ok {
		return b.onErr(nil)
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
