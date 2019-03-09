// subjectsToCASE.go
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
//		go run subjectsToCASE.go <subjects basename file> <base URI to use>
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

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
	Keyword                 string `csv:"Keyword"`
	NodeLineage             string `csv:"Node Lineage"`
}

// CFItems struct
type CFItems struct {
	CFItems map[string]CFItem `json:"CFItems"`
}

// CFItem struct
type CFItem struct {
	URI                string   `json:"uri"`
	HumanCodingScheme  string   `json:"humanCodingScheme"`
	CFDocumentURI      string   `json:"CFDocumentURI"`
	Identifier         string   `json:"identifier"`
	FullStatement      string   `json:"fullStatement"`
	ConceptKeywords    string   `json:"conceptKeywords"`
	EducationLevel     []string `json:"educationLevel"`
	LastChangeDateTime string   `json:"lastChangeDateTime"`
}

// LinkGenURI struct
type LinkGenURI struct {
	Title      string `json:"title"`
	Identifier string `json:"identifier"`
	URI        string `json:"uri"`
}

// CFAssociation struct
/*
{
	"uri": "http://frameworks.act.org/uri/8e87f5ca-3bbc-11e9-8a0b-dd5472351291",
	"identifier": "8e87f5ca-3bbc-11e9-8a0b-dd5472351291",
	"lastChangeDateTime": "2019-03-01T00:54:24+00:00",
	"CFDocumentURI": {
	  "title": "ACT Holistic Framework - Mathematics V2.0",
	  "identifier": "8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f",
	  "uri": "http://frameworks.act.org/uri/8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f"
	},
	"originNodeURI": {
	  "title": "H.A.MATH.NQ",
	  "identifier": "01befe44-bc4c-11e8-a572-0242ac120003",
	  "uri": "http://frameworks.act.org/uri/01befe44-bc4c-11e8-a572-0242ac120003"
	},
	"associationType": "isChildOf",
	"destinationNodeURI": {
	  "title": "ACT Holistic Framework - Mathematics V2.0",
	  "identifier": "8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f",
	  "uri": "http://frameworks.act.org/uri/8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f"
	},
	"sequenceNumber": 1
  },
*/
type CFAssociation struct {
	URI                string     `json:"uri"`
	Identifier         string     `json:"identifier"`
	CFDocumentURI      LinkGenURI `json:"CFDocumentURI"`
	OriginNodeURI      LinkGenURI `json:"originNodeURI"`
	DestinationNodeURI LinkGenURI `json:"destinationNodeURI"`
	AssociationType    string     `json:"associationType"`
	SequenceNumber     int        `json:"sequenceNumber"`
}

// CFAssociations struct
type CFAssociations struct {
	CFAssociations []CFAssociation
}

func (m *CFItems) loadSubjects(subjects Subjects, uriPrefix string, cfDocumentURI string) int {
	m.CFItems = make(map[string]CFItem)
	for _, v := range subjects.Subjects {
		var cfItem CFItem
		id, err := uuid.NewV4()
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error}).Info("No GUID for subject")
		}

		cfItem.Identifier = id.String()
		if len(v.Name) < 1 {
			v.Name = "(Unknown)"
		}
		const MaxHumanCodingScheme = 10
		if len(v.Name) > MaxHumanCodingScheme {
			cfItem.HumanCodingScheme = v.Name[0 : MaxHumanCodingScheme-1]
		} else {
			cfItem.HumanCodingScheme = v.Name
		}
		cfItem.FullStatement = v.Name
		cfItem.URI = uriPrefix + "/" + id.String()
		cfItem.CFDocumentURI = cfDocumentURI
		t := time.Now()
		cfItem.LastChangeDateTime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d+00:00",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		cfItem.EducationLevel = make([]string, 0)
		cfItem.ConceptKeywords = ""
		m.CFItems[v.Identifier] = cfItem
	}
	return len(m.CFItems)
}

// uses the supplementary subject info in the CSV to add additional metadata on each CFItem
// currently populating ConceptKeywords and EducationLevel
func (m *CFItems) loadSubjectsInfo(subjectsInfo SubjectsInfo, subjects Subjects) {
	var i CFItem
	for _, v := range subjectsInfo.SubjectsInfo {
		i = m.CFItems[v.NodeID]
		if v.Grades != "" {
			educationLevel := strings.Split(v.Grades, ",")
			if (len(educationLevel) > 0) {
				i.EducationLevel = educationLevel
			}
		}
		if v.Keyword != "" {
			i.ConceptKeywords = v.Keyword
		}
		m.CFItems[v.NodeID] = i
	}
}

// creates CASE isChildOf CFAssociations based on the parent-child relationships in the subject taxoonomy
// returns the number of associations created
// this is a sample top level isChildOf from ACT Holistic Framework Math
// NOTE THAT WE CREATE ISCHILDOF OF THE CFDOCUMENT IDENTIFIER
/*
   {
     "uri": "http://frameworks.act.org/uri/8e87f5ca-3bbc-11e9-8a0b-dd5472351291",
     "identifier": "8e87f5ca-3bbc-11e9-8a0b-dd5472351291",
     "lastChangeDateTime": "2019-03-01T00:54:24+00:00",
     "CFDocumentURI": {
       "title": "ACT Holistic Framework - Mathematics V2.0",
       "identifier": "8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f",
       "uri": "http://frameworks.act.org/uri/8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f"
     },
     "originNodeURI": {
       "title": "H.A.MATH.NQ",
       "identifier": "01befe44-bc4c-11e8-a572-0242ac120003",
       "uri": "http://frameworks.act.org/uri/01befe44-bc4c-11e8-a572-0242ac120003"
     },
     "associationType": "isChildOf",
     "destinationNodeURI": {
       "title": "ACT Holistic Framework - Mathematics V2.0",
       "identifier": "8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f",
       "uri": "http://frameworks.act.org/uri/8b5f90a6-3bbc-11e9-9fd0-97d95d55a29f"
     },
     "sequenceNumber": 1
   },
*/
func (m *CFAssociations) loadChildren(subjects Subjects, cfItems CFItems, uriPrefix string, cfDocument CFDocument) int {
	for _, v := range subjects.Subjects {

		var cfDocumentURI LinkGenURI
		cfDocumentURI.Identifier = cfDocument.Identifier
		cfDocumentURI.Title = cfDocument.Title
		cfDocumentURI.URI = cfDocument.URI
		var orgURI LinkGenURI
		var orgItem CFItem
		orgItem = cfItems.CFItems[v.Identifier]
		orgURI.Identifier = orgItem.Identifier
		orgURI.URI = uriPrefix + "/" + orgURI.Identifier
		orgURI.Title = cfItems.CFItems[v.Identifier].FullStatement
		var destURI LinkGenURI
		var destItem CFItem
		if v.Parent == "0" {
			destURI.Identifier = cfDocument.Identifier
			destURI.URI = cfDocument.URI
			destURI.Title = cfDocument.Title
		} else {
			destItem = cfItems.CFItems[v.Parent]
			destURI.Identifier = destItem.Identifier
			destURI.URI = uriPrefix + "/" + destURI.Identifier
			destURI.Title = cfItems.CFItems[v.Parent].FullStatement
		}

		var cfAssociation CFAssociation

		id, err := uuid.NewV4()
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error}).Info("Can't get UUID for association")
		}
		cfAssociation.URI = uriPrefix + "/cfAssociations/" + id.String()
		cfAssociation.Identifier = id.String()
		cfAssociation.CFDocumentURI = cfDocumentURI
		cfAssociation.OriginNodeURI = orgURI
		cfAssociation.DestinationNodeURI = destURI
		cfAssociation.AssociationType = "isChildOf"
		// TODO: compute sequence number
		//cfAssociation.sequenceNumber = 1
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

	id, err := uuid.NewV4()
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error}).Info("No GUID for subject")
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
	CFDocument     CFDocument        `json:"CFDocument"`
	CFItems        map[string]CFItem `json:"CFItems"`
	CFAssociations []CFAssociation   `json:"CFAssociations"`
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
		log.WithFields(log.Fields{"file": subjectsJSONFileName}).Fatal("Can't find specified file")
	} else {
		log.WithFields(log.Fields{"file": subjectsJSONFileName}).Info("Opened file")
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	// read our opened jsonfile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var subjects Subjects
	// we unmarshal our byteArray which contains our
	// jsonFile's content into subjects which we defined above
	json.Unmarshal(byteValue, &subjects)

	// second argument is base URI to be used for URI identifiers
	// default if nto provided is frameworks.act.org
	baseURI := "http://frameworks.act.org"
	if len(os.Args) > 2 {
		baseURI = os.Args[2]
	}
	caseSuffix := "/ims/case/v1p0" // current version of the  case spec
	uriPrefix := baseURI + caseSuffix
	var cfDocument CFDocument
	cfDocument.Init(uriPrefix, baseName)
	var cfItems CFItems
	count := cfItems.loadSubjects(subjects, uriPrefix, cfDocument.URI)
	log.WithFields(log.Fields{"count": count}).Info("Loaded subjects")

	// load supplementary information about subjects from CSV file
	subjectsInfoFileName := baseName + ".csv"
	subjectsInfoFile, err := os.OpenFile(subjectsInfoFileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"file": subjectsInfoFileName}).Error("Failed to open supplementary file")
	} else {
		defer subjectsInfoFile.Close()
		var subjectsInfo SubjectsInfo
		if err := gocsv.UnmarshalFile(subjectsInfoFile, &subjectsInfo.SubjectsInfo); err != nil { // Load subjects info from file
			log.WithFields(log.Fields{"file": subjectsInfoFileName}).Error("Failed to parse supplementary file")
		} else {
			log.WithFields(log.Fields{"count": count}).Info("Loaded supplementary subject info")
			// use the subjectsInfo array (from .csv file)
			// to populate the cfItems conceptKeywords and educationLevel fields
			cfItems.loadSubjectsInfo(subjectsInfo, subjects)
		}
	}

	// now use the parent child relationships in the subjects.json
	// to generate CASE "isChildOf" associations
	var cfAssociations CFAssociations
	cfAssociations.loadChildren(subjects, cfItems, uriPrefix, cfDocument)

	var cfPackage CFPackage
	cfPackage.CFDocument = cfDocument
	cfPackage.CFItems = cfItems.CFItems
	cfPackage.CFAssociations = cfAssociations.CFAssociations

	caseJSONFileName := baseName + "_case.json"
	caseJSON, err := json.MarshalIndent(cfPackage, "", "    ")
	err = ioutil.WriteFile(caseJSONFileName, caseJSON, 0644)
	if err != nil {
		log.WithFields(log.Fields{"file": caseJSONFileName}).Fatal("Can't write CASE file")
	} else {
		log.WithFields(log.Fields{"file": caseJSONFileName}).Info("Stored CFPackage to file")
	}
}
