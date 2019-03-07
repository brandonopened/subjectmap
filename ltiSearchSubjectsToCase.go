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
	URI                string `json:"uri"`
	HumanCodingScheme  string `json:"humanCodingScheme"`
	CFDocumentURI      string `json:"CFDocumentURI"`
	Identifier         string `json:"identifier"`
	LastChangeDateTime string `json:"lastChangeDateTime"`
}

// LinkGenURI struct
type LinkGenURI struct {
	Title      string `json:"title"`
	Identifier string `json:"identifier"`
	URI        string `json:"uri"`
}

// CFAssociation struct
type CFAssociation struct {
	OriginNodeURI      LinkGenURI `json:"originNodeURI"`
	DestinationNodeURI LinkGenURI `json:"destinationNodeURI"`
	AssociationType    string	  `json:"associationType"`
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
		cfItem.LastChangeDateTime = time.Now().Format("YYYY-MM-DDThh:mm:ss")
		m.CFItems[v.Identifier] = cfItem
	}
	return len(m.CFItems)
}

// FindOldID method
func (m *CFItems) FindOldID(oldID string) string {
	for k, v := range m.CFItems {
		if oldID == k {
			//fmt.Printf("Comparing %s: %s\n", k, oldID)
			return v.Identifier
		}
	}
	return ""
}

// FindOldParent method
func (m *CFItems) FindOldParent(oldParent string) string {
	for k, v := range m.CFItems {
		if oldParent == k {
			//fmt.Printf("Comparing %s to %s\n", k, oldParent)
			return v.Identifier
		}
	}
	return ""
}

func (m *CFAssociations) loadChildren(subjects Subjects, cfItems CFItems, uriPrefix string) int {
	for _, v := range subjects.Subjects {
		var orgURI LinkGenURI
		var destURI LinkGenURI
		orgURI.Identifier = cfItems.FindOldID(v.Identifier)
		orgURI.URI =  uriPrefix + "/" + orgURI.Identifier
		orgURI.Title = v.Name
		destURI.Identifier = cfItems.FindOldParent(v.Parent)
		destURI.URI = uriPrefix + "/" + destURI.Identifier
		destURI.Title = v.Name
		var cfAssociation CFAssociation
		cfAssociation.OriginNodeURI = orgURI
		cfAssociation.DestinationNodeURI = destURI
		cfAssociation.AssociationType = "isChildOf"
		m.CFAssociations = append(m.CFAssociations, cfAssociation)
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
	cfAssociations.loadChildren(subjects, cfItems, uriPrefix)

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
