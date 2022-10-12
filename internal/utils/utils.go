package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Trim remove all empty spaces from string
func Trim(str string) string {
	return regexp.MustCompile(`[\t\r\n ]+`).ReplaceAllString(str, "")
}

// Transforms a SQL result in JSON
// Reference https://github.com/elgs/gosqljson/blob/master/gosqljson.go
func QueryToJSONString(dbConn *sql.DB, sqlStatement string) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	results := []map[string]string{}
	rows, err := dbConn.Query(sqlStatement)
	if err != nil {
		fmt.Println("Error executing: ", sqlStatement)
		return "", err
	}
	cols, _ := rows.Columns()
	lenCols := len(cols)

	for i, v := range cols {
		cols[i] = strings.ToLower(v)
	}

	rawResult := make([][]byte, lenCols)

	dest := make([]any, lenCols) // A temporary any slice
	for i := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		result := make(map[string]string, lenCols)
		rows.Scan(dest...)
		for i, raw := range rawResult {
			if raw == nil {
				result[cols[i]] = ""
			} else {
				result[cols[i]] = string(raw)
			}
		}
		results = append(results, result)
	}

	jsonString, err := json.Marshal(results)
	return string(jsonString), err
}
