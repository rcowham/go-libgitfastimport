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
	"strconv"
)

type ezfiw struct {
	fiw fiWriter
	err error
}

func (e *ezfiw) WriteLine(a ...interface{}) {
	if e.err == nil {
		e.err = e.fiw.WriteLine(a...)
	}
}

func (e *ezfiw) WriteData(data string) {
	if e.err == nil {
		e.err = e.fiw.WriteData(data)
	}
}

func (e *ezfiw) WriteMark(idnum int) {
	if e.err == nil {
		e.err = e.fiw.WriteLine("mark", ":"+strconv.Itoa(idnum))
	}
}

type ezfir struct {
	fir fiReader
	err error
}

func (e *ezfir) Defer() error {
	if r := recover(); r != nil {
		if e.err != nil {
			return e.err
		}
		panic(r)
	}
	return nil
}

func (e *ezfir) Errcheck(err error) {
	if err == nil {
		return
	}
	e.err = err
	panic("everything is fine")
}

func (e *ezfir) PeekLine() string {
	line, err := e.fir.PeekLine()
	e.Errcheck(err)
	return line
}

func (e *ezfir) ReadLine() string {
	line, err := e.fir.ReadLine()
	e.Errcheck(err)
	return line
}
