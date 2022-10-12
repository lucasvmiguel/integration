package assertion

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	db, err := connectToDatabase()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	seed(db)
}

func TestSQLSetup_Success(t *testing.T) {
	assertion := SQLAssertion{}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_Success(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQLAssertion{
		DB: db,
		Query: `
		SELECT id, title, description, category_id FROM products
		`,
		ResultExpected: `
		[
			{"category_id":"1","description":"bar1","id":"1","title":"foo1"},
			{"category_id":"1","description":"bar2","id":"2","title":"foo2"}
		]
		`,
	}

	err := assertion.Assert()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_SuccessWithJoin(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQLAssertion{
		DB: db,
		Query: `
		SELECT products.id, products.title, products.category_id, categories.id, categories.name FROM products
		JOIN categories ON products.category_id = categories.id
		`,
		ResultExpected: `
		[
			{"category_id":"1","id":"1","name":"whatever","title":"foo1"},
			{"category_id":"1","id":"1","name":"whatever","title":"foo2"}
		]
		`,
	}

	err := assertion.Assert()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_FailedToQuery(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQLAssertion{
		DB: db,
		Query: `
		SELECT * FROM unknown
		`,
		ResultExpected: `
		[
			{"description":"bar1","id":"1","title":"foo1"},
			{"description":"bar2","id":"2","title":"foo2"}
		]
		`,
	}

	err := assertion.Assert()
	if err == nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_FailedResult(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQLAssertion{
		DB: db,
		Query: `
		SELECT * FROM products
		`,
		ResultExpected: `[]`,
	}

	err := assertion.Assert()
	if err == nil {
		t.Fatal(err)
	}
}

func connectToDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func seed(db *sql.DB) {
	db.Exec("DROP TABLE categories IF EXISTS;")
	db.Exec("DROP TABLE products IF EXISTS;")
	db.Exec("CREATE TABLE categories (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE);")
	db.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL UNIQUE, description TEXT NOT NULL, category_id int NOT NULL, FOREIGN KEY (category_id) REFERENCES categories (id) );")
	db.Exec("INSERT INTO categories (name) VALUES ('whatever');")
	db.Exec("INSERT INTO products (title, description, category_id) VALUES ('foo1', 'bar1', 1);")
	db.Exec("INSERT INTO products (title, description, category_id) VALUES ('foo2', 'bar2', 1);")
}
