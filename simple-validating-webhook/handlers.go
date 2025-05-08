package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	healthCheckMessage = map[string]string{"msg": "server is healthy"}
)

// healthcheck - Simple healthcheck that returns 200 ok
func (a *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthCheckMessage) // best‑effort; nothing we can do if this fails
}

// validate - Checks to see if the Kubernetes Pod object has the correct label
func (a *application) validate(w http.ResponseWriter, r *http.Request) {

	// Webhooks are sent a POST request, with Content-Type: application/json, with
	// an AdmissionReview API object in the admission.k8s.io API group serialized to JSON as the body.
	var input admissionv1.AdmissionReview
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		a.writeErrorMessage(w, "Unable to decode the POST request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// check for various nil or empty values
	if input.Request == nil || input.Request.RequestKind == nil {
		a.errorLog.Println("Request object is nil")
		a.writeErrorMessage(w, "invalid request", http.StatusBadRequest)
		return
	}

	// turn only for debugging
	// a.infoLog.Printf("%+v", input)

	// this webhook is only for Pod objects and this is to catch the misconfiguration of the webhook definition
	if input.Request.RequestKind.Kind != "Pod" {
		msg := fmt.Sprintf("Can not work with K8s %q objects, only with Pod", input.Request.RequestKind.Kind)
		a.writeErrorMessage(w, msg, http.StatusBadRequest)
		return
	}

	var (
		pod            corev1.Pod
		requestAllowed = false
		respMsg        = "Denied because the Pod is missing label " + a.cfg.Label
	)

	if len(input.Request.Object.Raw) <= 0 {
		a.writeErrorMessage(w, "empty Pod object in the request JSON", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(input.Request.Object.Raw, &pod); err != nil {
		a.writeErrorMessage(w, "Unable to marshal the raw payload into Pod object: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	// checking if the annotationKey "example.com/validate" exists with a value of true
	annotationExists, err := a.CheckNamespaceAnnotationTrue(a.cfg.Annotation, pod.Namespace)
	if err != nil {
		a.writeErrorMessage(w, "Unable to check annotations on the Pod "+err.Error(), http.StatusInternalServerError)
		return
	}

	// if the annotation Key was not preset or was set to false on the namespace
	// we have to skip the validation and allow the request
	if !annotationExists {
		a.infoLog.Printf("skipping validation of the Pod %s in namespace %s", pod.Name, pod.Namespace)
		respMsg = "skipping validation as annotation Key " + a.cfg.Annotation + " is missing or set to false on the namespace"
		a.craftAndWriteAdmissionResponse(w, input, respMsg, true)
		return
	}

	// if the annotationKey was present and is set to true
	// check if the Pod has the label matching a.cfg.Label, which is by default set to owner
	if len(pod.ObjectMeta.Labels) != 0 {
		if val, ok := pod.ObjectMeta.Labels[a.cfg.Label]; ok {
			if val != "" { // check if the value of the label is not empty
				requestAllowed = true
				respMsg = "Allowed as label " + a.cfg.Label + " is present in the Pod"
			}
			a.craftAndWriteAdmissionResponse(w, input, respMsg, requestAllowed)
			a.infoLog.Printf("\nAllowed Pod %q in namespace %q because label %q is present", pod.Name, pod.Namespace, a.cfg.Label)
			return
		}
	}

	// if the Pod does not have the label, we deny the request
	a.craftAndWriteAdmissionResponse(w, input, respMsg, requestAllowed)
	a.infoLog.Printf("\nDenied Pod %q in namespace %q because label %q is missing", pod.Name, pod.Namespace, a.cfg.Label)

}

// craftAndWriteAdmissionResponse - Helper function to craft and write the AdmissionReview response
// This function is used to send the response back to the Kubernetes API server
func (a *application) craftAndWriteAdmissionResponse(
	w http.ResponseWriter,
	input admissionv1.AdmissionReview,
	msg string,
	requestAllowed bool,
) {
	// we craft our final response here, which is an AdmissionReview object
	// we set the correct fiels and update the message
	output := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: input.TypeMeta.APIVersion,
			Kind:       input.TypeMeta.Kind,
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     input.Request.UID,
			Allowed: requestAllowed,
			Result: &metav1.Status{
				Message: msg,
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(output)
	if err != nil {
		a.writeErrorMessage(w, "Unable to marshal the json object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(resp); err != nil {
		a.writeErrorMessage(w, "Unable to send HTTP response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
