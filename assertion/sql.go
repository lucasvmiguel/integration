package assertion

import (
	"database/sql"
	"fmt"

	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/pkg/errors"
)

// SQL asserts a SQL query
type SQL struct {
	// DB database used to query the data to assert
	DB *sql.DB
	// Query that will run in the database
	Query call.Query
	// Result expects result in json that will be returned when the query run
	Result expect.Result
}

// Setup does not do anything because it doesn't need
func (a *SQL) Setup() error {
	return nil
}

// Assert checks if query returns the expected result
// Reference: https://kylewbanks.com/blog/query-result-to-map-in-golang
func (a *SQL) Assert() error {
	err := a.validate()
	if err != nil {
		return fmt.Errorf("failed to validate assertion: %w", err)
	}

	result := []map[string]interface{}{}
	rows, err := a.DB.Query(a.Query.Statement, a.Query.Params...)
	if err != nil {
		return fmt.Errorf("failed to execute SQL query: %w", err)
	}
	cols, _ := rows.Columns()

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		result = append(result, m)
	}

	numResults := len(result)
	numExpectedResult := len(a.Result)
	if numResults != numExpectedResult {
		return fmt.Errorf("SQL results don't match, it should have %d rows but it got %d rows", numExpectedResult, numResults)
	}

	for i, r := range result {
		for key, val := range r {
			if fmt.Sprint(a.Result[i][key]) != fmt.Sprint(val) {
				return fmt.Errorf("SQL result number %d don't match, it should be %v but it got %v", i, a.Result[i], r)
			}
		}
	}

	return nil
}

func (a *SQL) validate() error {
	if a.DB == nil {
		return errors.New("database is required")
	}

	if a.Query.Statement == "" {
		return errors.New("query statement is required")
	}

	if a.Result == nil {
		return errors.New("result is required")
	}

	return nil
}
