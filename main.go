package main

import (
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "os"

	"github.com/subtlepseudonym/notes/cmd"
)

const (
	notesDir       = "/Users/cdemille/workspace/log"
	filenameFormat = "%d.log"
)

type globalMeta struct {
	NumNotes  int `json:"numNotes"`
	IDCounter int `json:"idCounter"`
}

type noteMeta struct {
	Title    string `json:"title"`
	Created  int    `json:"created"`
	Modified int    `json:"modified"`
}

type note struct {
	Created int    `json:"created"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

func main() {
	// f, err := os.Open(fmt.Sprintf("%s/meta.log", notesDir))
	// if err != nil {
	// 	panic(err)
	// }

	// decoder := json.NewDecoder(f)

	// var global globalMeta
	// err = decoder.Decode(&global)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("global meta: %+v\n", global)

	// var meta []noteMeta
	// err = decoder.Decode(&meta)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, note := range meta {
	// 	fmt.Printf("%+v\n", note)
	// }

	// files, err := ioutil.ReadDir(notesDir)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, f := range files {
	// 	fmt.Printf("%d\t| %d\t| %s\n", f.Size(), f.ModTime().Unix(), f.Name())
	// }

	cmd.Execute()
}
