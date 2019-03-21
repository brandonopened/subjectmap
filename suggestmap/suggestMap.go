// subjectMap.go
//
// Given a set of items (skills/standards) loaded into OpenSALT
// Returns a list of items (skills) by similarity to a supplied phrase
//
// Generates <basename>_case.json file
//
// Usage:
//		go run subjectMap.go <phrase>
//
package main

import (
	"database/sql"
	"encoding/csv"
	"os"
	"strconv"
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
	sql := "SELECT identifier,uri, human_coding_scheme, full_statement, concept_keywords FROM ls_item WHERE full_statement LIKE '%" + phrase + "%';"
	log.WithFields(log.Fields{"SQL": sql}).Info("Query to execute")
	results, err := db.Query(sql)
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error()}).Fatal("Can't query!") // proper error handling instead of panic in your app
	}

	var matches []CFItem
	for results.Next() {
		var cfItem CFItem
		// for each row, scan the result into our tag composite object

		err = results.Scan(&cfItem.Identifier, &cfItem.URI, &cfItem.HumanCodingScheme, &cfItem.FullStatement, &cfItem.ConceptKeywords)
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error()}).Fatal("Failed to scan") // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		log.WithFields(log.Fields{"Identifier": cfItem.Identifier, "URI": cfItem.URI, "FullStatement": cfItem.FullStatement}).Info("Relevant to phrase")
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

func cfItemByIdentifier(db *sql.DB, identifier string) CFItem {
	// Execute the query
	sql := "SELECT identifier,uri, human_coding_scheme, full_statement, concept_keywords FROM ls_item WHERE identifier= '" + identifier + "';"
	log.WithFields(log.Fields{"SQL": sql}).Info("Query to execute")
	results, err := db.Query(sql)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Fatal("Can't query!") // proper error handling instead of panic in your app
	}

	var cfItem CFItem
	results.Next()
	err = results.Scan(&cfItem.Identifier, &cfItem.URI, &cfItem.HumanCodingScheme, &cfItem.FullStatement, &cfItem.ConceptKeywords)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed to scan") // proper error handling instead of panic in your app
	}
	// and then print out the tag's Name attribute
	log.WithFields(log.Fields{"Identifier": cfItem.Identifier, "FullStatement": cfItem.FullStatement}).Info("Retrieved item")

	return cfItem
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func itemsToCSV(cfItems []CFItem, fileName string) {
	//for _, v := range cfItems {
	//	log.WithFields(log.Fields{"Relevance": v.Relevance, "URI": v.URI, "FullStatement": v.FullStatement, "Identifier": v.Identifier}).Info("Relevant to item")
	//}

	file, err := os.Create(fileName)
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Identifier", "Relevance", "FullStatement", "URI"}
	err = writer.Write(headers)
	checkError("Can't write headers", err)

	for _, cfItem := range cfItems {
		record := []string{cfItem.Identifier.String, strconv.Itoa(cfItem.Relevance), cfItem.FullStatement.String, cfItem.URI.String}
		if err := writer.Write(record); err != nil {
			checkError("Failed to write item", err)
		}
	}

}

func main() {

	log.SetFormatter(&log.JSONFormatter{})
	identifier := "test"
	// first argument is the identifier of the skill or subject in OpenSALT
	if len(os.Args) > 1 {
		identifier = os.Args[1]
	}
	outputFile := identifier + ".csv"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}
	db := connectDB()

	item := cfItemByIdentifier(db, identifier)
	var allMatches []CFItem
	words := strings.Split(item.FullStatement.String, " ")
	for _, v := range words {
		matches := getRelatedItems(db, v)
		allMatches = append(allMatches, matches...)
	}
	itemMap := cfItemSliceToMap(allMatches)
	allMatches = itemsByRelevance(itemMap, item.FullStatement.String)
	db.Close()
	itemsToCSV(allMatches, outputFile)
}
