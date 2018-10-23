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

type Cmd interface {
	fiCmdRead(fiReader) (Cmd, error)
	fiCmdWrite(fiWriter) error
	fiCmdClass() cmdClass
}
