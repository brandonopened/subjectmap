package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Subjects struct
type Subjects struct {
	Subjects []Subject `json:"subjects"`
}

// Subject struct
type Subject struct {
	Identifier string `json:"identifier"`
	Parent     string `json:"parent"`
	Name       string `json:"name"`
}

// CFItems struct
type CFItems struct {
	CFItems []CFItem
}

// CFItem struct
type CFItem struct {
	Identifier        string
	URI               string
	HumanCodingScheme string
}

func subjectsToCFItems(subjects Subjects) CFItems {
	var cfItems CFItems

	for _, v := range subjects.Subjects {
		item := CFItem{Identifier: v.Identifier, HumanCodingScheme: v.Name}
		cfItems.CFItems = append(cfItems.CFItems, item)
	}
	return cfItems
}

func main() {
	var err error
	// Open our jsonFile
	baseName := "math"
	if len(os.Args) > 1 {
		baseName = os.Args[1]
	}
	fileName := baseName + ".json"
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Successfully opened %s\n", fileName)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var subjects Subjects

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &subjects)
	fmt.Printf("Finished loading %d subjects\n", len(subjects.Subjects))

	var cfItems CFItems
	cfItems = subjectsToCFItems(subjects)
	fmt.Printf("%d CFItems loaded\n", len(cfItems.CFItems))

	jsonPayload := []byte("INIT")
	jsonPayload, err = json.Marshal(cfItems)
	s := string(jsonPayload[:])
	fmt.Printf("JSON payload: %s\n", s)
}
