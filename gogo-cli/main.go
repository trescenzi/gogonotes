package main

import (
	"fmt"
	"os"
	"os/exec"
	//"text/template/parse"
	"regexp"
	"sort"
	"strconv"

	//"encoding/json"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getNoteRoot() string {
	var noteRoot string
	var ok bool
	if noteRoot, ok = os.LookupEnv("GOGONOTES_ROOT"); !ok {
		homeDir, err := os.UserHomeDir()
		handleErr(err)
		noteRoot = homeDir + "notes/"
	}
	err := os.MkdirAll(noteRoot, os.ModePerm)
	handleErr(err)
	return noteRoot
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func idFromNoteName(name string) int {
	re := regexp.MustCompile(`((\w*)-)?(\d*)(\.ggn)?`)
	match := re.FindAllStringSubmatch(name, 1)
	if len(match) == 0 {
		return -1
	}
	id, err := strconv.Atoi(match[0][3])
	if err != nil {
		return -1
	}
	return id
}

var noteRoot = getNoteRoot()

func main() {
	switch os.Args[1] {
	case "download":
		download(noteRoot, graphqlClient)
	case "save":
		id := idFromNoteName(os.Args[2])
		if id == -1 {
			err := fmt.Errorf("Please provide a note ID for saving")
			handleErr(err)
		}
		save(os.Args[2], id)
	case "new":
		notePrefix := ""
		if len(os.Args) == 3 {
			notePrefix = os.Args[2] + "-"
		}
		download(noteRoot, graphqlClient)
		f, err := os.Open(noteRoot)
		handleErr(err)
		names, err := f.Readdirnames(-1)
		handleErr(err)
		sort.Slice(names, func(i, j int) bool { return idFromNoteName(names[i]) < idFromNoteName(names[j]) })
		nextId := idFromNoteName(names[len(names)-1]) + 1
		notePath := noteRoot + notePrefix + fmt.Sprint(nextId) + ".ggn"
		os.Create(notePath)
		editorPath, err := exec.LookPath("nvim")
		handleErr(err)
		cmd := exec.Command(editorPath, notePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		handleErr(err)
		save(notePrefix + fmt.Sprint(nextId), nextId)
	default:
		err := fmt.Errorf("Options are download and save <id>")
		handleErr(err)
	}
}
//go:generate go run github.com/Khan/genqlient genqlient.yaml
