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
	"bufio"
	"io"
	"os"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type UnsupportedCommand string

func (e UnsupportedCommand) Error() string {
	return "Unsupported command: " + string(e)
}

// A Frontend is something that produces a fast-import stream; the
// Frontend object provides methods for reading from it.
type Frontend struct {
	fastImport   *parser
	catBlobWrite *textproto.CatBlobWriter
	catBlobFlush *bufio.Writer

	onErr func(error) error
}

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

func (f *Frontend) ReadCmd() (Cmd, error) {
	cmd, err := f.fastImport.ReadCmd()
	if err != nil {
		err = f.onErr(err)
	}
	return cmd, err
}

func (f *Frontend) RespondGetMark(sha1 string) error {
	err := f.catBlobWrite.WriteLine(sha1)
	if err != nil {
		return err
	}
	return f.catBlobFlush.Flush()
}

func (f *Frontend) RespondCatBlob(sha1 string, data string) error {
	err := f.catBlobWrite.WriteBlob(sha1, data)
	if err != nil {
		return err
	}
	return f.catBlobFlush.Flush()
}

func (f *Frontend) RespondLs(mode textproto.Mode, dataref string, path textproto.Path) error {
	var err error
	if mode == 0 {
		err = f.catBlobWrite.WriteLine("missing", path)
	} else {
		var t string
		switch mode {
		case textproto.ModeDir:
			t = "tree"
		case textproto.ModeGit:
			t = "commit"
		default:
			t = "blob"
		}
		err = f.catBlobWrite.WriteLine(mode, t, dataref+"\t"+textproto.PathEscape(path))
	}
	if err != nil {
		return err
	}
	return f.catBlobFlush.Flush()
}
