package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	//"text/template/parse"
	"regexp"
	"strconv"
	"sort"

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

func contains(s []string, searchterm string) bool {
    i := sort.SearchStrings(s, searchterm)
    return i < len(s) && s[i] == searchterm
}

func getTags(note string, existingTags []string) []string {
	re := regexp.MustCompile("#([A-Za-z0-9-_]+)")
	submatches := re.FindAllStringSubmatch(note, -1)
	var tags []string
	for _, sub := range submatches {
		if (sub[1] != "" && !contains(existingTags, sub[1])) {
      tags = append(tags, sub[1])
    }
	}
	return tags
}

func getLinks(note string, existingLinks []string) []int {
	re := regexp.MustCompile(`\[\[(\d+)\]\]`)
	links := re.FindAllStringSubmatch(note, -1)
	var linkIds []int
	for _, link := range links {
		if link[1] != "" && !contains(existingLinks, link[1]) {
			linkInt, err := strconv.Atoi(link[1])
			if err != nil {
				handleErr(fmt.Errorf("Links must be note IDs(ints) got" + link[1]))
			}
			linkIds = append(linkIds, linkInt)
		}
	}
	return linkIds
}

func createLinkAndTagInputsFromNote(note string, id int, existingTags []string, existingLinks []string) ([]*Note_tags_insert_input, []*Note_links_insert_input) {
	tags := getTags(note, existingTags)
	var tagsInput []*Note_tags_insert_input
	for _, tag := range tags {
		tagsInput = append(tagsInput, &Note_tags_insert_input{
			Note_id: id,
			Tag:     tag,
		})
	}
	links := getLinks(note, existingLinks)
	var linksInput []*Note_links_insert_input
	for _, link := range links {
		linksInput = append(linksInput, &Note_links_insert_input{
			From: id,
			To:   link,
		})
	}

	return tagsInput, linksInput;
}

func idFromNoteName(name string) int {
	re := regexp.MustCompile(`(\d+)\.ggn`)
	match := re.FindAllStringSubmatch(name, 1)
	if len(match) == 0 {
		return -1
	}
	id, err := strconv.Atoi(match[0][1])
	if err != nil {
		return -1
	}
	return id
}

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

func main() {

	noteRoot := getNoteRoot()
	graphqlClient := createGQLClient()

	switch os.Args[1] {
	case "download":
		download(noteRoot, graphqlClient)
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
			tagsInput, linksInput := createLinkAndTagInputsFromNote(newNote.Note, id, make([]string, 0), make([]string, 0))
			_, err = addNoteTagsAndLinks(context.Background(), graphqlClient, tagsInput, linksInput)
			handleErr(err)
			fmt.Println("Success! Saved new note " + fmt.Sprint(id))
		} else {
			//update existing note
			_, err := updateNote(context.Background(), graphqlClient, id, string(file))
			handleErr(err)

			rawExistingTags := notes.Notes[0].Note_tags;
			var existingTags []string
			for _, tag := range rawExistingTags {
				existingTags = append(existingTags, tag.Tag)
			}
			rawExistingLinks := notes.Notes[0].Note_links;
			var existingLinks []string
			for _, link := range rawExistingLinks {
				existingLinks = append(existingLinks, fmt.Sprint(link.To))
			}
			tagsInput, linksInput := createLinkAndTagInputsFromNote(string(file), id, existingTags, existingLinks)
			if len(tagsInput) == 0  && len(linksInput) == 0{
				fmt.Println("Success! Updated Note " + fmt.Sprint(id))
				return
			} else if len(tagsInput) == 0 && len(linksInput) != 0 {
				for _, link := range linksInput {
					fmt.Printf("Adding Link %d ", link.To)
				}
				fmt.Println()
				_, err = addNoteLinks(context.Background(), graphqlClient, linksInput)
				handleErr(err)
			} else if len(tagsInput) != 0 && len(linksInput) == 0 {
				for _, tag := range tagsInput {
					fmt.Printf("Adding Tag %s ", tag.Tag)
				}
				fmt.Println()
				_, err = addNoteTags(context.Background(), graphqlClient, tagsInput)
				handleErr(err)
			} else {
				for _, link := range linksInput {
					fmt.Printf("Adding Link %d ", link.To)
				}
				fmt.Println()
				for _, tag := range tagsInput {
					fmt.Printf("Adding Tag %s ", tag.Tag)
				}
				fmt.Println()
				_, err = addNoteTagsAndLinks(context.Background(), graphqlClient, tagsInput, linksInput)
				handleErr(err)
			}
			fmt.Println("Success! Updated Note " + fmt.Sprint(id))
		}
	case "new":
		download(noteRoot, graphqlClient)
		f, err := os.Open(noteRoot)
		handleErr(err)
		names, err := f.Readdirnames(-1)
		handleErr(err)
		sort.Slice(names, func(i, j int) bool { return idFromNoteName(names[i]) < idFromNoteName(names[j]) })
		nextId := idFromNoteName(names[len(names) - 1]) + 1
		os.Create(noteRoot + fmt.Sprint(nextId) + ".ggn")
		fmt.Println("Created Note " + fmt.Sprint(nextId))
	default:
		err := fmt.Errorf("Options are download and save <id>")
		handleErr(err)
	}
}

//go:generate go run github.com/Khan/genqlient genqlient.yaml
