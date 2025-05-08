package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHTTPRoutesUsingCorrectHTTPMethods(t *testing.T) {

	tt := []struct {
		name           string
		path           string
		wantStatusCode int
	}{
		{name: "test /healthcheck with GET method",
			path:           "/healthcheck",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "test /healthz with GET method",
			path:           "/healthz",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "test /thisdoesnotexist with GET method",
			path:           "/thisdoesnotexist",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "test /validate with POST method",
			path:           "/validate", // this exists but only POST method is allowed
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(fmt.Sprintf("testing route %v with HTTP GET method", tc.path), func(t *testing.T) {
			t.Parallel()
			app := &application{
				errorLog: log.New(io.Discard, "", log.Ldate),
				infoLog:  log.New(io.Discard, "", log.Ldate),
				cfg:      &envConfig{},
			}
			//srv := httptest.NewServer(app.routes())
			srv := httptest.NewServer(app.setupRoutes())
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

// TestHTTPRoutePOSTMethod - Test with valid Pod namespace, valid Admission review object
func TestHTTPRoutePOSTMethod(t *testing.T) {

	// create a fake Kubernetes client
	// this is then used to create a namespace object and query the same
	client := fake.NewSimpleClientset()

	app := &application{
		errorLog: log.New(io.Discard, "", log.Ldate),
		infoLog:  log.New(io.Discard, "", log.Ldate),
		cfg: &envConfig{
			Label:      "owner",
			Annotation: "example.com/validate",
		},
		client: client,
	}

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "webhook-demo", // this should match the namespace in the various json files in ./test-files
			Annotations: map[string]string{
				"example.com/validate": "true",
			},
		},
	}

	// Create the namespace using the fake client
	// namespace has correct annotation to force the validation
	_, err := client.CoreV1().Namespaces().Create(context.Background(), &ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatal("error creating namespace using fakeclient", err)
	}

	srv := httptest.NewServer(app.setupRoutes())
	defer srv.Close()
	url := fmt.Sprintf("%s/validate", srv.URL)

	testCase := map[string]struct {
		testPayloadFile string
		wantAllowed     bool
		wantStatus      int
	}{
		"valid request with missing owner label should be denied": {
			testPayloadFile: "test-files/admission-request-missing-labels.json",
			wantAllowed:     false,
			wantStatus:      http.StatusOK,
		},
		"valid request with owner label": {
			testPayloadFile: "test-files/admission-request-with-labels.json",
			wantAllowed:     true,
			wantStatus:      http.StatusOK,
		},
	}

	for name, tc := range testCase {
		tc := tc
		t.Run(name, func(t *testing.T) {

			f, err := os.Open(tc.testPayloadFile)
			if err != nil {
				t.Fatal(err)
			}
			// Post the AdmissionReview object to the /validate endpoint
			// and check the response
			resp, err := http.Post(url, "application/json", f)
			if err != nil {
				t.Fatal(err)
			}

			body, err := io.ReadAll(resp.Body)

			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			// uncomment when debugging
			// t.Logf("\n%s\n", string(body))

			output := &admissionv1.AdmissionReview{}

			// Unmarshal the response body into the AdmissionReview object
			// as the invalid-request.json contains a Pod without correct  request, we expect the response to be
			if err := json.Unmarshal(body, output); err != nil {
				t.Fatalf("error unmarshaling response body into AdmissionReview object - %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("\nHTTP post with valid AdmissionReview failed, returned status code was %v, want=%v",
					resp.StatusCode, tc.wantStatus)
				return
			}

			if output.Response.Allowed != tc.wantAllowed {
				t.Errorf("Admission response - want=%v, got=%v", output.Response.Allowed, tc.wantAllowed)
				return
			}

		})
	}

}
