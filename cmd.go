// Copyright (C) 2017-2018  Luke Shumaker <lukeshu@lukeshu.com>
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

// Package libfastimport implements reading and writing of git
// fast-import streams.
//
// The documentation here focuses on use of the package itself; it
// generally assumes a working understanding of the format.
// Documentation on the format itself can be found in the
// git-fast-import(1) man-page.
//
// A program can write commands to a backend by wrapping the
// appropriate io.Writer with a Backend object.
//
// A program can read commands from a frontend by wrapping the
// appropriate io.Reader with a Frontend object.
package libfastimport

type fiReader interface {
	PeekLine() (string, error)
	ReadLine() (string, error)
}

type fiWriter interface {
	WriteData(string) error
	WriteLine(a ...interface{}) error
}

type cmdClass int

const (
	cmdClassCommand cmdClass = 1 // may be a top-level command
	cmdClassCommit  cmdClass = 2 // may be used within in a commit

	cmdClassComment cmdClass = cmdClassCommand | cmdClassCommit
)

// Cmd is a command that may be found in a fast-import stream.
type Cmd interface {
	fiCmdRead(fiReader) (Cmd, error)
	fiCmdWrite(fiWriter) error
	fiCmdClass() cmdClass
}
