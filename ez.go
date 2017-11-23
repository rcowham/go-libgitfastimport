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
