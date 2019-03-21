// subjectMap.go
//
// Given a set of items (skills/standards) loaded into OpenSALT
// Returns a list of items by similarity to a supplied phrase
//
// Generates <basename>_case.json file
//
// Usage:
//		go run subjectMap.go <phrase>
//
package main

import (
	"database/sql"
	"os"
	"strings"

	slice "github.com/bradfitz/slice"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	smetrics "github.com/xrash/smetrics"
)

func connectDB() (conn *sql.DB) {
	db, err := sql.Open("mysql", "cftf:cftf@/cftf")
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error}).Error("Failed to open database")
	} else {
		log.WithFields(log.Fields{"DB": db.Driver}).Info("Opened database")
	}

	return db
}

// CFItem struct
type CFItem struct {
	URI                sql.NullString `json:"uri"`
	HumanCodingScheme  sql.NullString `json:"humanCodingScheme"`
	CFDocumentURI      sql.NullString `json:"CFDocumentURI"`
	Identifier         sql.NullString `json:"identifier"`
	FullStatement      sql.NullString `json:"fullStatement"`
	ConceptKeywords    sql.NullString `json:"conceptKeywords"`
	EducationLevel     sql.NullString `json:"educationLevel"`
	LastChangeDateTime sql.NullString `json:"lastChangeDateTime"`
	Relevance          int
}

func getRelatedItems(db *sql.DB, phrase string) []CFItem {

	// Execute the query
	sql := "SELECT uri, human_coding_scheme, full_statement, concept_keywords FROM ls_item WHERE full_statement LIKE '%" + phrase + "%';"
	log.WithFields(log.Fields{"SQL": sql}).Info("Query to execute")
	results, err := db.Query(sql)
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error}).Fatal("Can't query!") // proper error handling instead of panic in your app
	}

	var matches []CFItem
	for results.Next() {
		var cfItem CFItem
		// for each row, scan the result into our tag composite object

		err = results.Scan(&cfItem.URI, &cfItem.HumanCodingScheme, &cfItem.FullStatement, &cfItem.ConceptKeywords)
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error}).Fatal("Failed to scan") // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		log.WithFields(log.Fields{"URI": cfItem.URI, "FullStatement": cfItem.FullStatement}).Info("matches phrase")
		matches = append(matches, cfItem)
	}
	return matches
}

func scoreRelevance(cfItem CFItem, phrase string) int {
	score := smetrics.Ukkonen(cfItem.FullStatement.String, phrase, 1, 1, 2)
	return score
}

func itemsByRelevance(matches map[string]CFItem, phrase string) []CFItem {
	var result []CFItem
	for _, v := range matches {
		v.Relevance = scoreRelevance(v, phrase)
		result = append(result, v)
	}
	slice.Sort(result, func(i, j int) bool {
		return result[i].Relevance > result[j].Relevance
	})
	return result
}

func cfItemSliceToMap(elements []CFItem) map[string]CFItem {
	elementMap := make(map[string]CFItem)
	for _, item := range elements {
		elementMap[item.URI.String] = item
	}
	return elementMap
}

func main() {

	log.SetFormatter(&log.JSONFormatter{})
	phrase := "test"
	// first argument is basename of the JSON file with subjects
	if len(os.Args) > 1 {
		phrase = os.Args[1]
	}
	db := connectDB()

	var allMatches []CFItem
	words := strings.Split(phrase, " ")
	for _, v := range words {
		matches := getRelatedItems(db, v)
		allMatches = append(allMatches, matches...)
	}
	itemMap := cfItemSliceToMap(allMatches)
	allMatches = itemsByRelevance(itemMap, phrase)
	db.Close()
	for _, v := range allMatches {
		log.WithFields(log.Fields{"Relevance": v.Relevance, "URI": v.URI, "FullStatement": v.FullStatement, "Phrase": phrase}).Info("Matches phrase")
	}
}
