package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

func connectDB() (conn *sql.DB) {
	db, err := sql.Open("mysql", "cftf:cftf@/cftf")
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error}).Error("Failed to open database")
	} else {
		log.WithFields(log.Fields{"db": db.Driver}).Info("Opened database")
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
}

func getRelatedItems(db *sql.DB, phrase string) {
	// Execute the query
	sql := "SELECT uri, human_coding_scheme, full_statement, concept_keywords FROM ls_item WHERE full_statement LIKE '%" + phrase + "%';"
	log.WithFields(log.Fields{"SQL": sql}).Info("Query to execute")
	results, err := db.Query(sql)
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error}).Fatal("Can't query!") // proper error handling instead of panic in your app
	}
	for results.Next() {
		var cfItem CFItem
		// for each row, scan the result into our tag composite object

		err = results.Scan(&cfItem.URI, &cfItem.HumanCodingScheme, &cfItem.FullStatement, &cfItem.ConceptKeywords)
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error}).Fatal("Failed to scan") // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		log.WithFields(log.Fields{"URI": cfItem.URI, "FullStatement": cfItem.FullStatement}).Info("matches phrase")

	}

}

func main() {

	log.SetFormatter(&log.JSONFormatter{})
	phrase := "test"
	// first argument is basename of the JSON file with subjects
	if len(os.Args) > 1 {
		phrase = os.Args[1]
	}
	db := connectDB()
	getRelatedItems(db, phrase)
	db.Close()
}
