type FileAction interface {
	fiWrite(fi *FastImport) error
}

type FileModify struct {
	Mode    FileMode
	Path    string
	DataRef string
}

func (o FileModify) fiWrite(fi *FastImport) error {
	return fi.printf(w, "M %06o %s %s\n", o.Mode, o.DataRef, pathEscape(o.Path))
}

type FileModifyInline struct {
	Mode FileMode
	Path string
	Data []byte
}

func (o FileModifyInline) fiWrite(fi *FastImport) error {
	fi.printf("M %06o inline %s\n", o.Mode, pathEscape(o.Path))
	return fi.data(o.Data)
}

type FileDelete struct {
	Path string
}

func (o FileDelete) fiWrite(fi *FastImport) error {
	return fi.printf("D %s\n", pathEscape(o.Path))
}

type FileCopy struct {
	Src string
	Dst string
}

func (o FileCopy) fiWrite(fi *FastImport) error {
	return fi.printf("C %s %s\n", pathEscape(o.Src), pathEscape(o.Dst))
}

type FileRename struct {
	Src string
	Dst string
}

func (o FileRename) fiWrite(fi *FastImport) error {
	return fi.printf("R %s %s\n", pathEscape(o.Src), pathEscape(o.Dst))
}

type FileDeleteAll struct{}

func (o FileDeleteAll) fiWrite(fi *FastImport) error {
	return fi.printf("deleteall\n")
}

type NoteModify struct {
	CommitIsh string
	DataRef   string
}

func (o NoteModify) fiWrite(fi *FastImport) error {
	return fi.printf("N %s %s\n", o.DataRef, o.CommitIsh)
}

type NoteModifyInline struct {
	CommitIsh string
	Data      []byte
}

func (o NoteModify) fiWrite(fi *FastImport) error {
	fi.printf("N inline %s\n", o.CommitIsh)
	return fi.data(o.Data)
}

