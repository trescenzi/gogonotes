package main
import (
	"os"
	"strconv"
	"fmt"
	"context"
	"regexp"
	"time"
)

func save(id int) {
	filepath := noteRoot + fmt.Sprint(id) + ".ggn"
	fmt.Println("Saving " + filepath)
	file, err := os.ReadFile(filepath)
	handleErr(err)
	var notes *getNoteByIdResponse
	notes, err = getNoteById(context.Background(), graphqlClient, id)
	handleErr(err)
	if len(notes.Notes) == 0 {
		saveNewNote(file, id)
	} else {
		//update existing note
		_, err := updateNote(context.Background(), graphqlClient, id, string(file), time.Now())
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
		potentiallyAddTagsAndLinks(tagsInput, linksInput)
		fmt.Println("Success! Updated Note " + fmt.Sprint(id))
	}
}

func saveNewNote(file []byte, id int) {
	//add a new note
	creator, err := os.Hostname()
	handleErr(err)
	now := time.Now()
	newNote := &Notes_insert_input{
		Note:    string(file),
		Creator: creator,
		Id:      id,
		Created_at: now,
		Updated_at: now,
	}
	savedNote, err := addNote(context.Background(), graphqlClient, newNote)
	handleErr(err)
	fmt.Println(savedNote.Insert_notes_one.Id)
	tagsInput, linksInput := createLinkAndTagInputsFromNote(newNote.Note, id, make([]string, 0), make([]string, 0))
	_, err = addNoteTagsAndLinks(context.Background(), graphqlClient, tagsInput, linksInput)
	handleErr(err)
	fmt.Println("Success! Saved new note " + fmt.Sprint(id))
	potentiallyAddTagsAndLinks(tagsInput, linksInput)
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

func potentiallyAddTagsAndLinks(tagsInput []*Note_tags_insert_input, linksInput []*Note_links_insert_input) {
	if len(tagsInput) == 0  && len(linksInput) == 0 {
		return
	} else if len(tagsInput) == 0 && len(linksInput) != 0 {
		for _, link := range linksInput {
			fmt.Printf("Adding Link %d ", link.To)
		}
		fmt.Println()
		_, err := addNoteLinks(context.Background(), graphqlClient, linksInput)
		handleErr(err)
	} else if len(tagsInput) != 0 && len(linksInput) == 0 {
		for _, tag := range tagsInput {
			fmt.Printf("Adding Tag %s ", tag.Tag)
		}
		fmt.Println()
		_, err := addNoteTags(context.Background(), graphqlClient, tagsInput)
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
		_, err := addNoteTagsAndLinks(context.Background(), graphqlClient, tagsInput, linksInput)
		handleErr(err)
	}
}
