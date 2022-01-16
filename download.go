package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"github.com/Khan/genqlient/graphql"
)

func notePathIfExists(files []string, id int) (string, bool) {
	i := sort.Search(len(files), func(i int) bool { return idFromNoteName(files[i]) == id })
	if i < len(files) && idFromNoteName(files[i]) == id {
		return files[i], true
	}
	return "", false
}

func download(noteRoot string, graphqlClient graphql.Client) {
	var notes *GetAllNotesResponse
	notes, err := GetAllNotes(context.Background(), graphqlClient)
	if err != nil {
		return
	}

	f, err := os.Open(noteRoot)
	handleErr(err)
	names, err := f.Readdirnames(-1)
	handleErr(err)
	sort.Slice(names, func(i, j int) bool { return idFromNoteName(names[i]) < idFromNoteName(names[j]) })

	fmt.Println("Downloading ...")
	for _, note := range notes.Notes {
		filePath, alreadyExists := notePathIfExists(names, note.Id)
		if !alreadyExists {
			filePath = fmt.Sprint(note.Id) + ".ggn"
		}
		f, err := os.OpenFile(noteRoot + filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		handleErr(err)

		fStat, err := f.Stat()
		handleErr(err)
		buf := make([]byte, fStat.Size())

		fmt.Print(noteRoot + filePath)
		_, err = f.Read(buf)
		handleErr(err)

		f.WriteString(note.Note)
		f.Close()
	}
}
