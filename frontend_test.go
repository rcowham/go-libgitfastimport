// Tests for frontend

package libfastimport

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
author Robert Cowham <rcowham@perforce.com> 1644399073 +0000
committer Robert Cowham <rcowham@perforce.com> 1644399073 +0000
data 5
test
M 100644 :1 test.txt

`

	buf := strings.NewReader(input)
	f := NewFrontend(buf, nil, nil)
	for {
		cmd, err := f.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		switch cmd.(type) {
		case CmdBlob:
			blob := cmd.(CmdBlob)
			assert.Equal(t, 1, blob.Mark)
		case CmdReset:
			r := cmd.(CmdReset)
			assert.Equal(t, "refs/heads/main", r.RefName)
		case CmdCommit:
			c := cmd.(CmdCommit)
			assert.Equal(t, "refs/heads/main", c.Ref)
			assert.Equal(t, "Robert Cowham", c.Author.Name)
			assert.Equal(t, "rcowham@perforce.com", c.Author.Email)
			assert.Equal(t, "test\n", c.Msg)
		case CmdCommitEnd:
		case FileModify:
			f := cmd.(FileModify)
			assert.Equal(t, "test.txt", f.Path.String())
			assert.Equal(t, "100644", f.Mode.String())
			assert.Equal(t, ":1", f.DataRef)
		case FileDelete, FileCopy, FileRename:
			t.Error("Unexpected")
		default:
			t.Errorf("Unexpected cmd: %+v\n", cmd)
		}
	}

}
