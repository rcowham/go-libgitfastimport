// Copyright (C) 2017-2018, 2021  Luke Shumaker <lukeshu@lukeshu.com>
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
	"bufio"
	"io"

	"github.com/pkg/errors"
	"github.com/rcowham/go-libgitfastexport/textproto"
)

// A Backend is something that consumes a fast-import stream; the
// Backend object provides methods for writing to it.  A program that
// reads from a Backend would itself be a frontend.
//
// You may think of a "Backend" object as a "Writer" object, though it was not
// given that name because the GetMark, CatBlob, and Ls methods
// actually provide 2-way communication.
type Backend struct {
	fastImportClose io.Closer
	fastImportFlush *bufio.Writer
	fastImportWrite *textproto.FIWriter
	catBlob         *textproto.CatBlobReader

	inCommit bool

	err   error
	onErr func(error) error
}

// NewBackend creates a new Backend object that writes to the given
// io.WriteCloser.
//
// Optionally, you may also provide an io.Reader that responses to
// "cat-blob", "get-mark", and "ls" commands can be read from.
//
// Optionally, you may also provide an onErr function that can be used
// to handle or transform errors when they are encountered.
func NewBackend(fastImport io.WriteCloser, catBlob io.Reader, onErr func(error) error) *Backend {
	ret := &Backend{}

	ret.fastImportClose = fastImport
	ret.fastImportFlush = bufio.NewWriter(fastImport)
	ret.fastImportWrite = textproto.NewFIWriter(ret.fastImportFlush)

	if catBlob != nil {
		ret.catBlob = textproto.NewCatBlobReader(catBlob)
	}

	ret.onErr = func(err error) error {
		ret.err = err

		// Close the underlying writer, but don't let the
		// error mask the previous error.
		err = ret.fastImportClose.Close()
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

// Do tells the Backend to do the given command.
//
// It is an error (panic) if Cmd is a type that may only be used in a
// commit but we aren't in a commit.
func (b *Backend) Do(cmd Cmd) error {
	if b.err != nil {
		return b.err
	}

	switch {
	case !cmdIs(cmd, cmdClassInCommit):
		_, b.inCommit = cmd.(CmdCommit)
	case !b.inCommit && !cmdIs(cmd, cmdClassCommand):
		panic(errors.Errorf("Cannot issue commit sub-command outside of a commit: %[1]T(%#[1]v)", cmd))
	}

	err := cmd.fiCmdWrite(b.fastImportWrite)
	if err != nil {
		return b.onErr(err)
	}
	err = b.fastImportFlush.Flush()
	if err != nil {
		return b.onErr(err)
	}

	if _, isDone := cmd.(CmdDone); isDone {
		return b.onErr(nil)
	}

	return nil
}

// GetMark gets the SHA-1 referred to by the given mark from the
// Backend.
//
// It is an error (panic) to call GetMark if NewBackend did not have a
// cat-blob reader passed to it.
func (b *Backend) GetMark(cmd CmdGetMark) (sha1 string, err error) {
	err = b.Do(cmd)
	if err != nil {
		return
	}
	line, err := b.catBlob.ReadLine()
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

// CatBlob gets the SHA-1 and content of the specified blob from the
// Backend.
//
// It is an error (panic) to call CatBlob if NewBackend did not have a
// cat-blob reader passed to it.
func (b *Backend) CatBlob(cmd CmdCatBlob) (sha1 string, data string, err error) {
	err = b.Do(cmd)
	if err != nil {
		return
	}
	line, err := b.catBlob.ReadLine()
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

// Ls gets information about the file at the specified path from the
// Backend.
//
// It is an error (panic) to call Ls if NewBackend did not have a
// cat-blob reader passed to it.
func (b *Backend) Ls(cmd CmdLs) (mode Mode, dataref string, path Path, err error) {
	err = b.Do(cmd)
	if err != nil {
		return
	}
	line, err := b.catBlob.ReadLine()
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
