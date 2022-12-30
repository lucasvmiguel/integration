package integration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	goHTTP "net/http"
	"os"
	"testing"

	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"

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

func handlerCallHTTPPostJSON(w goHTTP.ResponseWriter, req *goHTTP.Request) {
	if req.Method != goHTTP.MethodPost {
		goHTTP.NotFound(w, req)
		return
	}

	body := struct {
		Title  string `json:"title"`
		UserID int    `json:"userId"`
	}{}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		goHTTP.Error(w, err.Error(), goHTTP.StatusBadRequest)
		return
	}

	if body.Title != "some title" || body.UserID != 1 {
		goHTTP.Error(w, "invalid body sent", goHTTP.StatusBadRequest)
		return
	}

	_, err = goHTTP.Post("https://jsonplaceholder.typicode.com/posts", "application/json", req.Body)
	if err != nil {
		goHTTP.Error(w, err.Error(), goHTTP.StatusInternalServerError)
		return
	}

	w.WriteHeader(goHTTP.StatusCreated)
	w.Write([]byte(`{
		"title": "some title",
		"description": "some description",
		"userId": 1,
		"comments": [
			{ "id": 1, "text": "foo" },
			{ "id": 2, "text": "bar" }
		]
	}`))
}

func handlerCallHTTPPost(w goHTTP.ResponseWriter, req *goHTTP.Request) {
	if req.Method != goHTTP.MethodPost {
		goHTTP.NotFound(w, req)
		return
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		goHTTP.Error(w, "failed to read request body", goHTTP.StatusBadRequest)
		return
	}

	expected := `
	hello
	world
	`

	if expected != string(reqBody) {
		goHTTP.Error(w, "invalid body sent", goHTTP.StatusBadRequest)
		return
	}

	_, err = goHTTP.Post("https://jsonplaceholder.typicode.com/posts", "application/json", req.Body)
	if err != nil {
		goHTTP.Error(w, err.Error(), goHTTP.StatusInternalServerError)
		return
	}

	respBody := `
	foo
	 bar`

	w.WriteHeader(goHTTP.StatusCreated)
	w.Write([]byte(respBody))
}

func init() {
	goHTTP.HandleFunc("/handlerCallHTTPGet", handlerCallHTTPGet)
	goHTTP.HandleFunc("/handlerCallHTTPPostJSON", handlerCallHTTPPostJSON)
	goHTTP.HandleFunc("/handlerCallHTTPPost", handlerCallHTTPPost)

	db, err := connectToDatabase()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	seed(db)

	go goHTTP.ListenAndServe(":8080", nil)
}

func TestHandlerCallHTTPPostJSON_Success(t *testing.T) {
	err := Test(&HTTPTestCase{
		Description: "TesthandlerCallHTTPPostJSON_Success",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPPostJSON",
			Method: goHTTP.MethodPost,
			Body: `{
				"title": "some title",
				"userId": 1
			}`,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusCreated,
			Body: `{
				"title": "some title",
				"description": "<<PRESENCE>>",
				"userId": 1,
				"comments": [
					{ "id": 1, "text": "foo" },
					{ "id": 2, "text": "bar" }
				]
			}`,
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts",
					Method: goHTTP.MethodPost,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPPost_Success(t *testing.T) {
	err := Test(&HTTPTestCase{
		Description: "TesthandlerCallHTTPPost_Success",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPPost",
			Method: goHTTP.MethodPost,
			Body: `
	hello
	world
	`,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusCreated,
			Body: `
	foo
	 bar`,
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts",
					Method: goHTTP.MethodPost,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_Success(t *testing.T) {

	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_Success",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
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

func TestHandlerCallHTTPGet_SuccessWithSQL(t *testing.T) {
	db, _ := connectToDatabase()
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_SuccessWithSQL",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
			&assertion.SQL{
				DB: db,
				Query: call.Query{
					Statement: `
					SELECT id, title, description, category_id FROM products
					`,
				},
				Result: expect.Result{
					{"id": 1, "title": "foo1", "description": "bar1", "category_id": 1},
					{"id": 2, "title": "foo2", "description": "bar2", "category_id": 1},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestHandlerCallHTTPGet_FailedMethod(t *testing.T) {
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_FailedMethod",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodPatch,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
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
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_FailedURL",
		Request: call.Request{
			URL:    "http://localhost:8080/invalid",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
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
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_WrongStatus",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusCreated,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
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
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_WrongResponseBody",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "invalid",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
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
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_InvalidAssertionHTTP",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
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

func TestHandlerCallHTTPGet_InvalidSQL(t *testing.T) {
	db, _ := connectToDatabase()
	err := Test(&HTTPTestCase{
		Description: "TestHandlerCallHTTPGet_Success",
		Request: call.Request{
			URL:    "http://localhost:8080/handlerCallHTTPGet",
			Method: goHTTP.MethodGet,
		},
		Response: expect.Response{
			StatusCode: goHTTP.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: goHTTP.MethodGet,
				},
			},
			&assertion.SQL{
				DB: db,
				Query: call.Query{
					Statement: "SELECT * FROM unknown",
				},
				Result: expect.Result{},
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
