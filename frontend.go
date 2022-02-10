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
	"os"

	"github.com/rcowham/go-libgitfastimport/textproto"
)

type UnsupportedCommand string

func (e UnsupportedCommand) Error() string {
	return "Unsupported command: " + string(e)
}

// A Frontend is something that produces a fast-import stream; the
// Frontend object provides methods for reading from it.  A program
// that writes to a Frontend would itself be a backend.
//
// You may think of a "Frontend" object as a "Reader" object, though
// it was not given that name because the RespondGetMark,
// RespondCatBlob, and RespondLs methods actually write information;
// it isn't a read-only object.
//
// The parser is a bit more forgiving than git's own parser.  It
// allows extra newlines anywhere it allows comments, whereas git's
// own parser is quite strict about newlines.  It allows 'cat-blob'
// and 'get-mark' commands anywhere it allows comments, which git used
// to allow, but was made stricter in git v2.22.0.
type Frontend struct {
	fastImport   *parser
	catBlobWrite *textproto.CatBlobWriter
	catBlobFlush *bufio.Writer

	onErr func(error) error
}

// NewFrontend creates a new Frontend object that reads from the given
// io.Reader.
//
// Optionally, you may also provide an io.Writer that responses to
// "cat-blob", "get-mark", and "ls" commands can be written to.
//
// Optionally, you may also provide an onErr function that can bue
// used to handle or transform errors when they are encountered.
func NewFrontend(fastImport io.Reader, catBlob io.Writer, onErr func(error) error) *Frontend {
	ret := &Frontend{}

	ret.fastImport = newParser(textproto.NewFIReader(fastImport))

	if catBlob == nil {
		catBlob = os.Stdout
	}
	ret.catBlobFlush = bufio.NewWriter(catBlob)
	ret.catBlobWrite = textproto.NewCatBlobWriter(ret.catBlobFlush)

	if onErr == nil {
		onErr = func(e error) error { return e }
	}
	ret.onErr = onErr

	return ret
}

// ReadCmd reads a command from the Frontend.
func (f *Frontend) ReadCmd() (Cmd, error) {
	cmd, err := f.fastImport.ReadCmd()
	if err != nil {
		err = f.onErr(err)
	}
	return cmd, err
}

// RespondGetMark sends to the Frontend a response to a "get-mark"
// command.
//
// It is an error (panic) to call RespondGetMark if NewFrontend did
// not have a cat-blob writer passed to it.
func (f *Frontend) RespondGetMark(sha1 string) error {
	err := f.catBlobWrite.WriteLine(sha1)
	if err != nil {
		return err
	}
	return f.catBlobFlush.Flush()
}

// RespondCatBlob sends to the Frontend a response to a "cat-blob"
// command.
//
// It is an error (panic) to call RespondCatBlob if NewFrontend did
// not have a cat-blob writer passed to it.
func (f *Frontend) RespondCatBlob(sha1 string, data string) error {
	err := f.catBlobWrite.WriteBlob(sha1, data)
	if err != nil {
		return err
	}
	return f.catBlobFlush.Flush()
}

// RespondLs sends to the Frontend a response to a "ls" command.
//
// It is an error (panic) to call RespondLs if NewFrontend did not
// have a cat-blob writer passed to it.
func (f *Frontend) RespondLs(mode Mode, dataref string, path Path) error {
	var err error
	if mode == 0 {
		err = f.catBlobWrite.WriteLine("missing", path)
	} else {
		var t string
		switch mode {
		case ModeDir:
			t = "tree"
		case ModeGit:
			t = "commit"
		default:
			t = "blob"
		}
		err = f.catBlobWrite.WriteLine(mode, t, dataref+"\t"+PathEscape(path))
	}
	if err != nil {
		return err
	}
	return f.catBlobFlush.Flush()
}
