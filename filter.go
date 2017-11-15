package main

import (
	"os"
	"os/exec"
	"bytes"
)

var prog [string]string

func init() {
	var err error
	for _, p := range []string{"git", "makepkg"} {
		prog[p], err = exec.LookPath("git")
		if err != nil {
			panic(err)
		}
	}
}

func pipeline(cmds ...*exec.Cmd) err error {
	for i, cmd := range cmds[:len(cmds)-1] {
		cmds[i+1].Stdin, err = cmd.StdoutPipe()
		if err != nil {
			return
		}
	}

	stderr := make([]bytes.Buffer, len(cmds))
	for i, cmd := range cmds {
		cmd.Stderr = &stderr[i]
		if err = cmd.Start(); err != nil {
			break
		}
	}

	for i, cmd := range cmds {
		if cmd.Process == nil {
			continue
		}
		if _err := cmd.Wait(); _err != nil {
			if ee, ok := _err.(*exec.ExitError); ok {
				ee.Stderr = stderr[i].Bytes()
			}
			if err != nil {
				err = _err
			}
		}
	}
	return
}

func pkgbuild2srcinfo(pkgbuildId string) (string, error) {
	cachefilename := "filter/pkgbuild2srcinfo/"+pkgbuildId
	for {
		b, err := ioutil.ReadFile(cachefilename)
		if err == nil && len(bytes.TrimSpace(b)) == 40 {
			return bytes.TrimSpace(b).String(), nil
		}
	
		file, err := os.OpenFile("filter/tmp/PKGBUILD", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return "", err
		}
		err = pipeline(
			&exec.Cmd{
				Path: prog["git"], Args: ["git", "cat-file", "blob", pkgbuildId],
				Stdout: file,
			})
		file.Close()
		if err != nil {
			return "", err
		}


		file, err = os.OpenFile(cachefilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		err = pipeline(
			&exec.Cmd{
				Path: prog["makepkg"], Args: ["makepkg", "--printsrcinfo"],
				Dir: "filter/tmp"
			},
			&exec.Cmd{
				Path: prog["git"], Args: ["git", "hash-object", "-t", "blob", "-w", "--stdin", "--no-filters"],
				Stdout: &buf,
			})
		file.Close()
		if err != nil {
			return "", err
		}
	}
}



func filter(fromPfix string, toPfix string) {
	exec.Cmd{
		Path: prog["git"], Args: ["git", "fast-export",
			"--use-done-feature",
			"--no-data",
			"--", fromPfix + "/master"],
	}
	
}
