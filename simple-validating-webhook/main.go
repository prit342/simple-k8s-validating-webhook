package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
)

// Application holds an instance of an application
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	cfg      *envConfig
}

// type envConfig holds various environment variables
type envConfig struct {
	CertPath string `env:"CERT_PATH" envDefault:"/source/cert.pem"`
	KeyPath  string `env:"KEY_PATH" envDefault:"/source/key.pem"`
	Port     int    `env:"PORT" envDefault:"3000"`
}

func main() {

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	cfg := envConfig{}

	if err := env.Parse(&cfg); err != nil {
		errorLog.Fatalln(err)
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		cfg:      &cfg,
	}

	tlsPair, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)

	if err != nil {
		app.errorLog.Fatalln("Error loading TLS certs", err)
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", cfg.Port), // Listen on all the interfaces
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsPair}},
	}

	server.Handler = app.routes()

	go func() {
		infoLog.Printf("Starting the web server on port %v", cfg.Port)
		app.errorLog.Println(server.ListenAndServeTLS("", ""))

	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	app.infoLog.Println("Got shutdown signal, shutting down the web server gracefully...")

	server.Shutdown(context.Background())

}
