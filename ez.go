package libfastimport

import (
	"strconv"

	"github.com/pkg/errors"
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

var ezPanic = errors.New("everything is fine")

func (e *ezfir) Defer() error {
	if e.err != nil {
		r := recover()
		if r == nil {
			panic("ezfir.err got set, but didn't panic")
		}
		if r != ezPanic {
			panic(r)
		}
		return e.err
	}
	return nil
}

func (e *ezfir) Errcheck(err error) {
	if err == nil {
		return
	}
	e.err = err
	panic(ezPanic)
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
