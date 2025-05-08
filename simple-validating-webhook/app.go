package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"k8s.io/client-go/kubernetes"
)

// Application holds an instance of an application
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	cfg      *envConfig
	client   kubernetes.Interface
}

// type envConfig holds various environment variables
type envConfig struct {
	CertPath   string `env:"CERT_PATH" envDefault:"/source/cert.pem"`
	KeyPath    string `env:"KEY_PATH" envDefault:"/source/key.pem"`
	Port       int    `env:"PORT" envDefault:"3000"`
	Annotation string `env:"ANNOTATION" envDefault:"example.com/validate"`
	Label      string `env:"LABEL" envDefault:"owner"`
}

// GetKubeConfig - return a valid kube config or an error
func GetKubeConfig() (*rest.Config, error) {

	var err error
	var config *rest.Config

	// Try reading kubeconfig file form the environment variable
	if envVar := os.Getenv("KUBECONFIG"); len(envVar) > 0 {
		if config, err = clientcmd.BuildConfigFromFlags("", envVar); err == nil {
			return config, nil
		}
	}

	errorMsg := "error loading kubeconfig from environment variable KUBECONFIG"

	// Try getting kube config form the home directory
	if home := homedir.HomeDir(); home != "" {

		kubeconfig := filepath.Join(home, ".kube", "config")
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err == nil {
			return config, nil
		}

	}

	errorMsg = errorMsg + "\n" + err.Error()

	// finally try to get in-cluster config via service account
	if config, err = rest.InClusterConfig(); err == nil {
		return config, nil
	}

	errorMsg = errorMsg + "\n" + err.Error()

	return config, fmt.Errorf("failed to authenticate with the Kubernets cluster - \n%v", errorMsg)
}

// NewKubeClient - returns a new kuberenets client set
func NewKubeClient(config *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

// CheckNamespaceAnnotationTrue - returns true if the value of an annotationKey is present and set to true on a namespace
func (app *application) CheckNamespaceAnnotationTrue(annotation, namespace string) (bool, error) {

	if app == nil || app.client == nil {
		return false, fmt.Errorf("application or client is nil")
	}

	ns, err := app.client.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})

	if err != nil {
		nsErr := fmt.Errorf("error checking annotations on the namespace %v - %v", namespace, err)
		app.errorLog.Println(nsErr)
		return false, nsErr
	}

	for key, val := range ns.GetAnnotations() {
		if key == annotation && strings.ToLower(val) == "true" {
			app.infoLog.Printf("Found annotationKey %v set to value %v in the namespace %v", annotation, val, namespace)
			return true, nil
		}
	}

	return false, nil
}
