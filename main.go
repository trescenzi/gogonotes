package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	//"text/template/parse"
	"strconv"

	//"encoding/json"

	"github.com/Khan/genqlient/graphql"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-hasura-admin-secret", t.key)
	return t.wrapped.RoundTrip(req)
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func createGQLClient() graphql.Client {
	key := os.Getenv("HASURA_ADMIN_SECRET")
	if key == "" {
		err := fmt.Errorf("must set HASURA_ADMIN_SECRET")
		handleErr(err)
	}

	endpoint := os.Getenv("HASURA_ENDPOINT")

	if endpoint == "" {
		err := fmt.Errorf("must set HASURA_ENDPOINT")
		handleErr(err)
	}

	httpClient := http.Client{
		Transport: &authedTransport{
			key:     key,
			wrapped: http.DefaultTransport,
		},
	}
	return graphql.NewClient(endpoint, &httpClient)
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

func main() {

	noteRoot := getNoteRoot()
	graphqlClient := createGQLClient()

	switch os.Args[1] {
	case "download":
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
	case "save":
		id, err := strconv.Atoi(os.Args[2])
		if id == 0 || err != nil {
			err := fmt.Errorf("Please provide a note ID for saving")
			handleErr(err)
		}

		filepath := noteRoot + fmt.Sprint(id) + ".ggn"
		fmt.Println("Saving " + filepath)
		file, err := os.ReadFile(filepath)
		handleErr(err)
		var notes *getNoteByIdResponse
		notes, err = getNoteById(context.Background(), graphqlClient, id)
		handleErr(err)

		if len(notes.Notes) == 0 {
			//add a new note
			creator, err := os.Hostname()
			handleErr(err)
			newNote := &Notes_insert_input{
				Note:    string(file),
				Creator: creator,
				Id:      id,
			}
			savedNote, err := addNote(context.Background(), graphqlClient, newNote)
			handleErr(err)
			fmt.Println(savedNote.Insert_notes_one.Id)
			fmt.Println("Success! Saved new note " + fmt.Sprint(id))
		} else {
			//update existing note
			_, err := updateNote(context.Background(), graphqlClient, id, string(file))
			handleErr(err)
			fmt.Println("Success! Updated Note " + fmt.Sprint(id))
		}
	default:
		err := fmt.Errorf("1 you fuck", os.Args[0])
		handleErr(err)
	}
}

//go:generate go run github.com/Khan/genqlient genqlient.yaml
