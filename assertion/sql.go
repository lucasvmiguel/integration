package assertion

import (
	"database/sql"
	"fmt"
	"integration/internal/utils"

	"github.com/pkg/errors"
)

// SQLAssertion asserts a SQL query
type SQLAssertion struct {
	// DB database used to query the data to assert
	DB *sql.DB
	// Query that will run in the database
	Query string
	// ResultExpected expects result in json that will be returned when the query run.
	// A multiline string is valid
	// eg:
	// [{
	// 		"description":"Bar",
	// 		"id":"2",
	// 		"title":"Fooa"
	// 	}]
	ResultExpected string
}

// Setup does not do anything because it doesn't need
func (a *SQLAssertion) Setup() error {
	return nil
}

// Assert checks if query returns the expected result
func (a *SQLAssertion) Assert() error {
	sqlResult, err := utils.QueryToJSONString(a.DB, a.Query)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to call database in sql operation: %s", a.Query))
	}

	resultTrim := utils.Trim(sqlResult)
	resultExpected := utils.Trim(a.ResultExpected)
	if resultTrim != resultExpected {
		return errors.Errorf("sql operation should be %s it got %s", resultExpected, resultTrim)
	}

	return nil
}
