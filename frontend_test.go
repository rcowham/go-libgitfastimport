// Tests for frontend

package libfastimport

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func writeToTempFile(contents string) string {
	f, err := os.CreateTemp("", "*.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprint(f, contents)
	if err != nil {
		fmt.Println(err)
	}
	return f.Name()
}

func TestParseBasic(t *testing.T) {
	input := `blob
mark :1
data 5
test

reset refs/heads/main
commit refs/heads/main
mark :2
committer John Smith <john@example.com> 1644399073 +0000
author John Smith <john@example.com> 1644399073 +0000
data 5
test
M 100644 :1 test.txt

`

	fname := writeToTempFile(input)
	file, err := os.Open(fname)
	if err != nil {
		fmt.Printf("ERROR: Failed to open file '%s': %v\n", fname, err)
		t.Fail()
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	f := NewFrontend(buf, nil, nil)
	for {
		cmd, err := f.ReadCmd()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		switch cmd.(type) {
		case CmdBlob:
			blob := cmd.(CmdBlob)
			fmt.Printf("Blob: Mark:%d OriginalOID:%s\n", blob.Mark, blob.OriginalOID)
		case CmdReset:
			reset := cmd.(CmdReset)
			fmt.Printf("Reset: - %+v\n", reset)
		case CmdCommit:
			commit := cmd.(CmdCommit)
			fmt.Printf("Commit:  %+v\n", commit)
		case CmdCommitEnd:
			commit := cmd.(CmdCommitEnd)
			fmt.Printf("CommitEnd:  %+v\n", commit)
		case FileModify:
			f := cmd.(FileModify)
			fmt.Printf("FileModify:  %+v\n", f)
		case FileDelete:
			f := cmd.(FileDelete)
			fmt.Printf("FileModify: Path:%s\n", f.Path)
		case FileCopy:
			f := cmd.(FileCopy)
			fmt.Printf("FileCopy: Src:%s Dst:%s\n", f.Src, f.Dst)
		case FileRename:
			f := cmd.(FileRename)
			fmt.Printf("FileRename: Src:%s Dst:%s\n", f.Src, f.Dst)
		default:
			fmt.Printf("Not handled\n")
			fmt.Printf("Found cmd %v\n", cmd)
			fmt.Printf("Cmd type %v\n", reflect.TypeOf(cmd))
		}
	}

}
