// subjectsToCase.go
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
//		go run subjectsToCase.go <subjects basename file> <base URI to use>
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gocarina/gocsv"
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

// SubjectsInfo struct
type SubjectsInfo struct {
	SubjectsInfo []SubjectInfo `json:"subjects"`
}

//SubjectInfo struct
type SubjectInfo struct {
	ParentNodeID            string `csv:"Parent Node ID"`
	NodeID                  string `csv:"Node ID"`
	NodeTitle               string `csv:"Node Title"`
	NodeResources           string `csv:"Node Resources"`
	HierarachyNodeResources string `csv:"Node Resources"`
	Grades                  string `csv:"Grades"`
	Keyword                 string `csv:"Keywords"`
	NodeLineage             string `csv:"Node Lineage"`
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
	ConceptKeywords    string `json:"conceptKeywords"`
	EducationLevel     string `json:"educationLevel"`
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
	AssociationType    string     `json:"associationType"`
}

// CFAssociations struct
type CFAssociations struct {
	CFAssociations []CFAssociation
}

func (m *CFItems) loadSubjects(subjects Subjects, uriPrefix string, cfDocumentURI string) int {
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
		cfItem.CFDocumentURI = cfDocumentURI
		cfItem.LastChangeDateTime = time.Now().Format("YYYY-MM-DDThh:mm:ss")
		m.CFItems[v.Identifier] = cfItem
	}
	return len(m.CFItems)
}

func (m *CFItems) loadSubjectsInfo(subjectsInfo SubjectsInfo, subjects Subjects) {
	var i CFItem 
	for _, v := range subjectsInfo.SubjectsInfo {
		i=m.CFItems[v.NodeID]
		i.ConceptKeywords = v.Keyword
		i.EducationLevel = v.Grades
		m.CFItems[v.NodeID]=i
	}
}

func (m *CFAssociations) loadChildren(subjects Subjects, cfItems CFItems, uriPrefix string) int {
	for _, v := range subjects.Subjects {
		var orgURI LinkGenURI
		var destURI LinkGenURI
		var orgItem CFItem
		orgItem = cfItems.CFItems[v.Identifier]
		orgURI.Identifier = orgItem.Identifier
		orgURI.URI = uriPrefix + "/" + orgURI.Identifier
		orgURI.Title = v.Name
		var destItem CFItem
		destItem = cfItems.CFItems[v.Parent]
		destURI.Identifier = destItem.Identifier
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
	m.Creator = "subjectsToCase"
	m.Title = baseName
	m.Description = baseName
	m.LastChangeDateTime = time.Now().Format("YYYY-MM-DD hh:mm:ss")
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

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var subjects Subjects
	// we unmarshal our byteArray which contains our
	// jsonFile's content into subjects which we defined above
	json.Unmarshal(byteValue, &subjects)

	// now set a base name for all the URI identifiers
	baseURI := "http://frameworks.act.org"
	if len(os.Args) > 2 {
		baseURI = os.Args[2]
	}
	caseSuffix := "/ims/case/v1/p0"
	uriPrefix := baseURI + caseSuffix
	var cfDocument CFDocument
	cfDocument.Init(uriPrefix, baseName)
	var cfItems CFItems
	count := cfItems.loadSubjects(subjects, uriPrefix, cfDocument.URI)
	fmt.Printf("Finished loading %d subjects\n", count)

	subjectsInfoFileName := baseName + ".csv"
	subjectsInfoFile, err := os.OpenFile(subjectsInfoFileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer subjectsInfoFile.Close()
	var subjectsInfo SubjectsInfo
	if err := gocsv.UnmarshalFile(subjectsInfoFile, &subjectsInfo.SubjectsInfo); err != nil { // Load subjects info from file
		panic(err)
	}

	// use the subjectsInfo array (from .csv file)
	// to populate the cfItems conceptKeywords and educationLevel fields
	cfItems.loadSubjectsInfo(subjectsInfo, subjects)

	// now use the parent child relationships in the subjects.json
	// to generate CASE "isChildOf" associations
	var cfAssociations CFAssociations
	cfAssociations.loadChildren(subjects, cfItems, uriPrefix)

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
