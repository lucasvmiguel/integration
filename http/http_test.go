package http

import (
	"database/sql"
	"fmt"
	"integration/assertion"
	goHTTP "net/http"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func handlerCallHTTPGet(w goHTTP.ResponseWriter, req *goHTTP.Request) {
	if req.Method != goHTTP.MethodGet {
		goHTTP.NotFound(w, req)
		return
	}

	_, err := goHTTP.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		goHTTP.Error(w, err.Error(), goHTTP.StatusInternalServerError)
	}

	fmt.Fprintf(w, "hello")
}

func init() {
	goHTTP.HandleFunc("/handlerCallHTTPGet", handlerCallHTTPGet)

	db, err := connectToDatabase()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	seed(db)

	go goHTTP.ListenAndServe(":8080", nil)
}

func TestHandlerCallHTTPGet_Success(t *testing.T) {

	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_Success",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_SuccessWithSQLAssertion(t *testing.T) {
	db, _ := connectToDatabase()
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_Success",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
			&assertion.SQLAssertion{
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
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_FailedMethod(t *testing.T) {
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_FailedMethod",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodPatch,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
		},
	})

	if err == nil {
		t.Fatal("it should return an error due to an invalid method")
	}
}

func TestHandlerCallHTTPGet_FailedURL(t *testing.T) {
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_FailedURL",
		Request: Request{
			URL:    "http://localhost:8080/invalid",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
		},
	})

	if err == nil {
		t.Fatal("it should return an error due to an invalid method")
	}
}

func TestHandlerCallHTTPGet_WrongStatus(t *testing.T) {
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_WrongStatus",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusCreated,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
		},
	})

	if err == nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_WrongResponseBody(t *testing.T) {
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_WrongResponseBody",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "invalid",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
		},
	})

	if err == nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_InvalidHTTPAssertion(t *testing.T) {
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_InvalidAssertionHTTP",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://invalid",
					Method: goHTTP.MethodGet,
				},
			},
		},
	})

	if err == nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_InvalidSQLAssertion(t *testing.T) {
	db, _ := connectToDatabase()
	err := Test(TestCase{
		Description: "TestHandlerCallHTTPGet_Success",
		Request: Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		ResponseExpected: Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: assertion.RequestExpected{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
			&assertion.SQLAssertion{
				DB: db,
				Query: `
				SELECT * FROM unknown
				`,
				ResultExpected: `[]`,
			},
		},
	})

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