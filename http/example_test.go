package http

import (
	"net/http"
	"testing"
)

func init() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	go http.ListenAndServe(":8080", nil)
}

func TestPingEndpoint(t *testing.T) {
	err := Test(TestCase{
		Description: "Testing ping endpoint",
		Request: Request{
			URL:    "http://localhost:8080/ping",
			Method: http.MethodGet,
		},
		ResponseExpected: ResponseExpected{
			StatusCode: http.StatusOK,
			Body:       "pong",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
