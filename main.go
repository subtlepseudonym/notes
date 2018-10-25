package main

import (
	"fmt"
	"os"
	"encoding/json"
	"time"
)

const notesDir = "/Users/cdemille/workspace/log"

type globalMeta struct {
	NumNotes int `json:"numNotes"`
}

type noteMeta struct {
	Title    string `json:"title"`
	Created  int    `json:"created"`
	Modified int    `json:"modified"`
}

func main() {
	f, err := os.Open(fmt.Sprintf("%s/meta.log", notesDir))
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(f)

	var global globalMeta
	err = decoder.Decode(&global)
	if err != nil {
		panic(err)
	}
	fmt.Printf("global meta: %+v\n", global)

	var meta []noteMeta
	err = decoder.Decode(&meta)
	if err != nil {
		panic(err)
	}
	for _, note := range meta {
		fmt.Printf("%+v\n", note)
	}

	fmt.Println(time.Now().Unix())
}
