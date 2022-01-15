package main

import (
	"context"
	"fmt"
	"os"
	"github.com/Khan/genqlient/graphql"
)

func download(noteRoot string, graphqlClient graphql.Client) {
	var notes *GetAllNotesResponse
	notes, err := GetAllNotes(context.Background(), graphqlClient)
	if err != nil {
		return
	}

	for _, note := range notes.Notes {
		var filePath = noteRoot + fmt.Sprint(note.Id) + ".ggn"
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
		handleErr(err)

		fStat, err := f.Stat()
		handleErr(err)
		buf := make([]byte, fStat.Size())

		fmt.Println("Downloading " + filePath)
		_, err = f.Read(buf)
		handleErr(err)

		f.WriteString(note.Note)
		f.Close()
	}
}
