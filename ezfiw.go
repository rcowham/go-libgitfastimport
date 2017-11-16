package libfastimport

import (
	"strconv"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type ezfiw struct {
	fiw *textproto.FIWriter
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
