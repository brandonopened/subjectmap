// ltiSearchSubjectsToCase.go
//
// Loads JSON file with a subject taxonomy expressed in IMS Global LTI Resource Search
// Subjects taxonomy payload format
//
// Saves as valid IMS Global CASE format, which can be loaded into any CASE compliant
// skill/standards manager
//
// Generates <basename>_case.json file
//
// Usage:
//		go run ltiSearchSubjectsToCase.go <subjects basename file> <base URI to use>
//

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

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
	CFItems map[string]CFItem `json:"CFItems"`
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
	Title      string
	Identifier string
	URI        string
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

func (m *CFItems) loadSubjects(subjects Subjects, uriPrefix string) int {
	m.CFItems = make(map[string]CFItem)
	for _, v := range subjects.Subjects {
		var cfItem CFItem
		id, err := uuid.NewV1()
		if err != nil {
			fmt.Println("Can't create GUID")
		}
		cfItem.Identifier = id.String()
		cfItem.HumanCodingScheme = v.Name
		cfItem.URI = uriPrefix + "/" + id.String()
		m.CFItems[v.Identifier] = cfItem
	}
	return len(m.CFItems)
}

func (m *CFAssociations) loadChildCFAssociations(subjects Subjects, uriPrefix string) int {
	for _, v := range subjects.Subjects {
		var orgURI, destURI LinkGenURI

		orgURI.Identifier = uuid.Must(uuid.NewV4()).String()
		orgURI.URI = uriPrefix + "/" + v.Identifier
		association := CFAssociation{OriginNodeURI: orgURI, DestinationNodeURI: destURI, AssociationType: isChildOf}
		m.CFAssociations = append(m.CFAssociations, association)
	}
	return len(m.CFAssociations)
}

// CFDocument struct
type CFDocument struct {
	URI                string `json:"uri"`
	CFPackageURI       string `json:"CFPackageURI"`
	Identifier         string `json:"identifier"`
	Creator            string `json:"creator"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	LastChangeDateTime string `json:"lastChangeDateTime"`
}

// Init ... set up the CFDocument structure
func (m *CFDocument) Init(uriPrefix string, baseName string) {

	id, err := uuid.NewV1()
	if err != nil {
		fmt.Println("Failed to create unique ID")
	}
	m.URI = uriPrefix + "/CFDocuments/" + id.String()
	m.CFPackageURI = uriPrefix + "/CFPackages/" + id.String()
	m.Identifier = id.String()
	m.Creator = "S2S"
	m.Title = baseName
	m.Description = baseName
	m.LastChangeDateTime = time.Now().Format("YYYY-MM-DDThh:mm:ss")
}

// CFPackage struct
type CFPackage struct {
	CFDocument     CFDocument     `json:"CFDocument"`
	CFItems        CFItems        `json:"CFItems"`
	CFAssociations CFAssociations `json:"CFAssociations"`
}

func main() {
	var err error

	baseName := "math"
	// first argument is basename of the JSON file with subjects
	if len(os.Args) > 1 {
		baseName = os.Args[1]
	}
	subjectsJSONFileName := baseName + ".json"
	// Open our jsonFile with the LTI Resource Search subject taxonomy payload
	jsonFile, err := os.Open(subjectsJSONFileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Successfully opened %s\n", subjectsJSONFileName)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// now set a base name for all the URI identifiers
	baseURI := "http://frameworks.act.org"
	if len(os.Args) > 2 {
		baseURI = os.Args[2]
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

	var cfItems CFItems
	count := cfItems.loadSubjects(subjects, uriPrefix)
	fmt.Printf("Finished loading %d subjects\n", count)
	var cfAssociations CFAssociations
	cfAssociations.loadChildCFAssociations(subjects, uriPrefix)

	var cfDocument CFDocument
	cfDocument.Init(uriPrefix, baseName)

	var cfPackage CFPackage
	cfPackage.CFDocument = cfDocument
	cfPackage.CFItems = cfItems
	cfPackage.CFAssociations = cfAssociations

	caseJSONFileName := baseName + "_case.json"
	caseJSON, err := json.Marshal(cfPackage)
	fmt.Println("Writing CASE to " + caseJSONFileName)
	err = ioutil.WriteFile(caseJSONFileName, caseJSON, 0644)
	if err != nil {
		panic(err)
	}
}
