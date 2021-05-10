package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHTTPRoutesUsingGetMethod(t *testing.T) {

	tt := []struct {
		path           string
		wantStatusCode int
	}{
		{
			path:           "/healthcheck",
			wantStatusCode: http.StatusOK,
		},
		{
			path:           "/thisdoesnotexist",
			wantStatusCode: http.StatusNotFound,
		},
		{
			path:           "/validate", // this exists but only POST method is allowed
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("testing route %v with HTTP GET method", tc.path), func(t *testing.T) {
			t.Parallel()
			app := &application{
				errorLog: log.New(ioutil.Discard, "", log.Ldate),
				infoLog:  log.New(ioutil.Discard, "", log.Ldate),
				cfg:      &envConfig{},
			}
			srv := httptest.NewServer(app.routes())
			defer srv.Close()

			res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, tc.path))

			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			if err != nil {
				t.Fatalf("Could not send GET request to %v; %v", tc.path, err)
			}

			if err != nil {
				t.Fatalf("Could not read response, %v", err)
			}

			got := res.StatusCode

			if got != tc.wantStatusCode {
				t.Errorf("call to HTTP route %v failed; HTTP status code - want=%v got=%v", tc.path, tc.wantStatusCode, got)
			}

		})
	}

}

func TestHTTPRoutePOSTMethod(t *testing.T) {

	app := &application{
		errorLog: log.New(ioutil.Discard, "", log.Ldate),
		infoLog:  log.New(ioutil.Discard, "", log.Ldate),
		cfg:      &envConfig{},
	}

	srv := httptest.NewServer(app.routes())
	defer srv.Close()

	f, err := os.Open("test-files/invalid-request.json")

	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("%s/validate", srv.URL)

	resp, err := http.Post(url, "application/json", f)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("HTTP post with valid AdmissionReview failed, returned status code was not 200OK")
	}

}
