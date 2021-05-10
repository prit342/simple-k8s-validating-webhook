package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
)

//func TestValidateWebHookWithInvalidInput(t *testing.T) {
//
//	app := &application{
//		errorLog: log.New(ioutil.Discard, "", log.Ldate),
//		infoLog:  log.New(ioutil.Discard, "", log.Ldate),
//		cfg:      &envConfig{},
//	}
//
//	f, err := os.Open("test-files/invalid-request.json")
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	rr := httptest.NewRecorder()
//
//	handler := http.HandlerFunc(app.validate)
//
//	// send Admission review loaded from the json file
//	// this instances of admission review is missing the required Pod label of owner
//	req, err := http.NewRequest("POST", "/validate", f)
//
//	if err != nil {
//		t.Fatal( err)
//	}
//
//	handler.ServeHTTP(rr, req)
//
//	result := admissionv1.AdmissionReview{}
//
//	err = json.NewDecoder(rr.Body).Decode(&result)
//
//	if err != nil {
//		t.Fatal( err)
//	}
//
//	want := false
//	got := result.Response.Allowed
//	if want != got {
//		t.Errorf("Admission review was not denied: want=%v, instead got=%v", want, got)
//	}
//}

//func TestValidateWebHookWithInvalidInput(t *testing.T) {
//
//	app := &application{
//		errorLog: log.New(ioutil.Discard, "", log.Ldate),
//		infoLog:  log.New(ioutil.Discard, "", log.Ldate),
//		cfg:      &envConfig{},
//	}
//
//	var arRequest admissionv1.AdmissionReview
//	var arResponse admissionv1.AdmissionReview
//
//	requestBytes, err := ioutil.ReadFile("test-files/invalid-request.json")
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if err:= json.Unmarshal(requestBytes, &arRequest); err != nil {
//		t.Fatal(err)
//	}
//
//	rr := httptest.NewRecorder()
//
//	handler := http.HandlerFunc(app.validate)
//
//	reader := bytes.NewReader(requestBytes)
//
//	req, err := http.NewRequest(http.MethodPost, "/validate", reader)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	handler.ServeHTTP(rr, req)
//
//	err = json.NewDecoder(rr.Body).Decode(&arResponse)
//
//	want := false
//	got := arResponse.Response.Allowed
//
//	if want != got {
//		t.Errorf("Admission review was not denied: want=%v, instead got=%v", want, got)
//	}
//
//
//}

func TestValidateWebhookHandler(t *testing.T) {

	tt := []struct {
		name           string
		allowed        bool
		sourceJsonFile string
	}{
		{
			name:           "Test webhook when the pod is missing label owner",
			allowed:        false,
			sourceJsonFile: "test-files/invalid-request.json",
		},
		{
			name:           "Test webhook when the pod is has correct labels",
			allowed:        true,
			sourceJsonFile: "test-files/valid-request.json",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := &application{
				errorLog: log.New(ioutil.Discard, "", log.Ldate),
				infoLog:  log.New(ioutil.Discard, "", log.Ldate),
				cfg:      &envConfig{},
			}
			f, err := os.Open(tc.sourceJsonFile)
			defer f.Close()
			if err != nil {
				t.Fatalf("Failed to load input json file %v", err.Error())
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(app.validate)

			// send Admission review loaded from the json file
			req, err := http.NewRequest("POST", "/validate", f)

			if err != nil {
				t.Fatalf("Failed to create the request object %v", err.Error())
			}

			handler.ServeHTTP(rr, req)

			result := admissionv1.AdmissionReview{}
			err = json.NewDecoder(rr.Body).Decode(&result)

			if err != nil {
				t.Fatalf("Failed to decode the Json response to AdmissionReview object %v", err.Error())
			}

			admissionReviewReqAllowed := result.Response.Allowed

			if admissionReviewReqAllowed != tc.allowed {
				t.Errorf("AdmissionReview.Request.Allowed field: want=%v got=%v", tc.allowed, admissionReviewReqAllowed)
			}

		})
	}

}
