// ltiSearchSubjectsToCase.go
//
// Loads JSON file with a subject taxonomy expressed in IMS Global LTI Resource Search 
// Subjects taxonomy payload format
//
// Saves as valid IMS Global CASE format, which can be loaded into any CASE compliant
// skill/standards manager
// 
// Usage:
//		go run ltiSearchSubjectsToCase.go <subjects json file> <base URI to use>

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/gofrs/uuid"
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

const ( // iota is reset to 0
	isChildOf     = iota
	isPeerOf      = iota
	isPartOf      = iota
	exactMatchOf  = iota
	precedes      = iota
	isRelatedTo   = iota
	replacedBy    = iota
	exemplar      = iota
	hasSkillLevel = iota
)


// LinkGenURI struct
type LinkGenURI struct {
	Title		string
	Identifier  string
	URI			string		
}

// CFAssociation struct
type CFAssociation struct {
	OriginNodeURI      LinkGenURI
	DestinationNodeURI LinkGenURI
	AssociationType    int
}

// CFAssociations struct
type CFAssociations struct {
	CFAssociations []CFAssociation
}


func subjectsToCFItems(subjects Subjects,uriPrefix string) CFItems {
	var cfItems CFItems

	for _, v := range subjects.Subjects {
		uri := uriPrefix + "/" + v.Identifier
		item := CFItem{Identifier: v.Identifier, HumanCodingScheme: v.Name, URI: uri}
		cfItems.CFItems = append(cfItems.CFItems, item)
	}
	return cfItems
}

func subjectsToChildCFAssociations(subjects Subjects,uriPrefix string) CFAssociations {
	var cfAssociations CFAssociations
	for _, v := range subjects.Subjects {
		var orgURI,destURI LinkGenURI

		orgURI.Identifier = uuid.Must(uuid.NewV4()).String()
		orgURI.URI = uriPrefix + "/" + v.Identifier
		association := CFAssociation{OriginNodeURI:orgURI,DestinationNodeURI:destURI,AssociationType:isChildOf}
		cfAssociations.CFAssociations=append(cfAssociations.CFAssociations,association)
	}
	return cfAssociations
}

func generateCASEFile(cfItems CFItems,cfAssociations CFAssociations) {
	// TODO: perform CASE generation
}

func main() {
	var err error

	baseName := "math"
	// first argument is basename of the JSON file with subjects
	if len(os.Args) > 1 {
		baseName = os.Args[1]
	}
	fileName := baseName + ".json"
	// Open our jsonFile with the LTI Resource Search subject taxonomy payload
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Successfully opened %s\n", fileName)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// now set a base name for all the URI identifiers
	baseURI := "http://frameworks.act.org"
	if len(os.Args) > 1 {
		baseName = os.Args[2]
	}
	caseSuffix := "/ims/case/v1/p0"
	uriPrefix := baseURI + caseSuffix

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var subjects Subjects

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &subjects)
	fmt.Printf("Finished loading %d subjects\n", len(subjects.Subjects))

	var cfItems CFItems
	cfItems = subjectsToCFItems(subjects,uriPrefix)
	fmt.Printf("%d CFItems loaded\n", len(cfItems.CFItems))
	jsonPayload := []byte("INIT")
	jsonPayload, err = json.Marshal(cfItems)
	s := string(jsonPayload[:])
	fmt.Printf("JSON payload: %s\n", s)

	var cfAssociations CFAssociations
	cfAssociations = subjectsToChildCFAssociations(subjects,uriPrefix)

	generateCASEFile(cfItems,cfAssociations)
}
