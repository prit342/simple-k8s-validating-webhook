package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Simple healthcheck that returns 200 ok
func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintf(w, "%s", `{"msg": "server is healthy"}`)
	if err != nil {
		app.errorLog.Println("error writing response", err)
		return
	}

}


// Checks to see if the Kubernetes object has the correct label
func (app *application) validate(w http.ResponseWriter, r *http.Request) {

	// Webhooks are sent a POST request, with Content-Type: application/json, with
	// an AdmissionReview API object in the admission.k8s.io API group serialized to JSON as the body.
	input := admissionv1.AdmissionReview{}

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		app.writeErrorMessage(w, "Unable to decode the POST request: "+err.Error())
		return
	}

	// This is to catch the misconfiguration of the webhook definition
	switch input.Request.RequestKind.Kind {
	case "Pod":
		app.infoLog.Println("Request came for object type of Pod")

		pod := corev1.Pod{}

		var requestAllowed = false
		var respMsg = "Denied because the Pod is missing label " + app.cfg.Label

		if err := json.Unmarshal(input.Request.Object.Raw, &pod); err != nil {
			app.writeErrorMessage(w, "Unable to marshal the raw payload into Pod object: "+err.Error())
			return
		}

		// checking if the annotationKey "example.com/validate" exists with a value of true
		ok, err := app.CheckNamespaceAnnotationTrue(app.cfg.Annotation, pod.Namespace)

		if err != nil {
			app.writeErrorMessage(w, "Unable to check annotationKey on the Pod "+err.Error())
			return
		}

		// if the annotationKey was not preset or was set to false
		if !ok {
			app.infoLog.Printf("skipping validation of the Pod %s in namespace %s", pod.Name, pod.Namespace)
			requestAllowed = true
			respMsg = "skipping validation as annotationKey " + app.cfg.Annotation + " is missing or set to false"
		}

		if ok && len(pod.ObjectMeta.Labels) > 0 {

			if val, ok := pod.ObjectMeta.Labels[app.cfg.Label]; ok {
				if val != "" {
					requestAllowed = true
					respMsg = "Allowed as label " + app.cfg.Label + " is present in the Pod"
				}
				app.infoLog.Printf("Allowed Pod %v in namespace %v because label %v is present", pod.Name, pod.Namespace, app.cfg.Label)
			}
		}

		output := admissionv1.AdmissionReview{

			Response: &admissionv1.AdmissionResponse{
				UID:     input.Request.UID,
				Allowed: requestAllowed,
				Result: &metav1.Status{
					Message: respMsg,
				},
			},
		}

		output.TypeMeta.Kind = input.TypeMeta.Kind
		output.TypeMeta.APIVersion = input.TypeMeta.APIVersion

		w.Header().Set("Content-Type", "application/json")

		resp, err := json.Marshal(output)

		if err != nil {
			app.writeErrorMessage(w, "Unable to marshal the json object: "+err.Error())
			return
		}

		if _, err := w.Write(resp); err != nil {
			app.writeErrorMessage(w, "Unable to send HTTP response: "+err.Error())
			return
		}

	default:
		msg := fmt.Sprintf("Can not work with K8s %v objects, only with Pod", input.Request.RequestKind.Kind)
		app.writeErrorMessage(w, msg)
	}

}
