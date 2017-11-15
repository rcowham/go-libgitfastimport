package libfastimport

type FileAction interface {
	fiWriteFA(*FIWriter) error
}

type FileModify struct {
	Mode    Mode
	Path    Path
	DataRef string
}

func (o FileModify) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("M", o.Mode, o.DataRef, o.Path)
}

type FileModifyInline struct {
	Mode Mode
	Path Path
	Data []byte
}

func (o FileModifyInline) fiWriteFA(fiw *FIWriter) error {
	ez := &ezfiw{fiw: fiw}
	ez.WriteLine("M", o.Mode, "inline", o.Path)
	ez.WriteData(o.Data)
	return ez.err
}

type FileDelete struct {
	Path Path
}

func (o FileDelete) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("D", o.Path)
}

type FileCopy struct {
	Src Path
	Dst Path
}

func (o FileCopy) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("C", o.Src, o.Dst)
}

type FileRename struct {
	Src string
	Dst string
}

func (o FileRename) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("R", o.Src, o.Dst)
}

type FileDeleteAll struct{}

func (o FileDeleteAll) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("deleteall")
}

type NoteModify struct {
	CommitIsh string
	DataRef   string
}

func (o NoteModify) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("N", o.DataRef, o.CommitIsh)
}

type NoteModifyInline struct {
	CommitIsh string
	Data      []byte
}

func (o NoteModifyInline) fiWriteFA(fiw *FIWriter) error {
	ez := &ezfiw{fiw: fiw}
	ez.WriteLine("N", "inline", o.CommitIsh)
	ez.WriteData(o.Data)
	return ez.err
}

// See CmdLs for using ls outside of a commit
type FileLs struct {
	Path Path
}

func (o FileLs) fiWriteFA(fiw *FIWriter) error {
	return fiw.WriteLine("ls", o.Path)
}
