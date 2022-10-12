package assertion

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
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
	assertion := SQL{}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_Success(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQL{
		DB: db,
		Query: call.Query{
			Statement: "SELECT id, title, description, category_id FROM products",
		},
		Result: expect.Result{
			{"id": 1, "title": "foo1", "description": "bar1", "category_id": 1},
			{"id": 2, "title": "foo2", "description": "bar2", "category_id": 1},
		},
	}

	err := assertion.Assert()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_SuccessWithJoin(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQL{
		DB: db,
		Query: call.Query{
			Statement: `
				SELECT products.id, products.title, products.category_id, categories.id as cat_id, categories.name FROM products
				JOIN categories ON products.category_id = categories.id
			`,
		},
		Result: expect.Result{
			{"id": 1, "title": "foo1", "category_id": 1, "name": "whatever", "cat_id": 1},
			{"id": 2, "title": "foo2", "category_id": 1, "name": "whatever", "cat_id": 1},
		},
	}

	err := assertion.Assert()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_FailedToQuery(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQL{
		DB: db,
		Query: call.Query{
			Statement: "SELECT * FROM unknown",
		},
		Result: expect.Result{
			{"id": 1, "title": "foo1", "description": "bar1", "category_id": 1},
			{"id": 2, "title": "foo2", "description": "bar2", "category_id": 1},
		},
	}

	err := assertion.Assert()
	if err == nil {
		t.Fatal(err)
	}
}

func TestSQLAssert_FailedResult(t *testing.T) {
	db, _ := connectToDatabase()
	assertion := SQL{
		DB: db,
		Query: call.Query{
			Statement: "SELECT * FROM products",
		},
		Result: expect.Result{},
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
