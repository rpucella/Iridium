package main

import (
	"io/ioutil"
	"path"
	"strings"
)

func getPassageNames(srcdir string) ([]string, error) { 
	files, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".txt") {
			names = append(names, strings.TrimSuffix(f.Name(), ".txt"))
		}
	}
	return names, nil
}

func readPassage(srcdir string, passage string) (string, error) {
	content, err := ioutil.ReadFile(path.Join(srcdir, passage + ".txt"))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writePassage(srcdir string, passage string, text string) (error) {
	return ioutil.WriteFile(path.Join(srcdir, passage + ".txt"), []byte(text), 0644)
}
