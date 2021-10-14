package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
)

func CreateNamespace(t *testing.T, namespace string, annotations map[string]string, client *fake.Clientset) {

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace,
			Annotations: annotations,
		},
	}

	ctx := context.Background()

	n, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	fmt.Println(n.Annotations)

	if err != nil {
		t.Fatal("error creating namespace", err)
	}

}

func TestValidateWebhookHandler(t *testing.T) {

	tt := []struct {
		name            string
		allowed         bool
		sourceJsonFile  string
		annotationkey   string
		annotationValue string
	}{
		{
			name:            "Pod is missing label owner and namespace has correct annotations",
			allowed:         false,
			sourceJsonFile:  "test-files/invalid-request.json",
			annotationkey:   "example.com/validate",
			annotationValue: "true",
		},
		{
			name:            "Pod has correct labels and namespace has correct annotations",
			allowed:         true,
			sourceJsonFile:  "test-files/valid-request.json",
			annotationkey:   "example.com/validate",
			annotationValue: "true",
		},
		{
			name:            "Pod has correct labels but annotation is set to false",
			allowed:         true,
			sourceJsonFile:  "test-files/valid-request.json",
			annotationkey:   "example.com/validate",
			annotationValue: "false",
		},
	}

	for _, tc := range tt {
		tc := tc // capture inner variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewSimpleClientset()

			app := &application{
				errorLog: log.New(ioutil.Discard, "", log.Ldate),
				infoLog:  log.New(ioutil.Discard, "", log.Ldate),
				cfg: &envConfig{
					Annotation: "example.com/validate",
					Label:      "owner",
				},
				client: client,
			}

			testAnnotations := make(map[string]string)

			testAnnotations[tc.annotationkey] = tc.annotationValue

			CreateNamespace(t, "webhook-demo", testAnnotations, client)

			f, err := os.Open(tc.sourceJsonFile)

			defer func() {
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			}()

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

			t.Log(result)

			admissionReviewReqAllowed := result.Response.Allowed

			if admissionReviewReqAllowed != tc.allowed {
				t.Errorf("AdmissionReview.Request.Allowed field: want=%v got=%v", tc.allowed, admissionReviewReqAllowed)
			}

		})
	}

}
