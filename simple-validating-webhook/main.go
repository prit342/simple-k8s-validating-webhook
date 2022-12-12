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

func main() {
	
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	
	cfg := envConfig{}
	
	if err := env.Parse(&cfg); err != nil {
		errorLog.Fatalln(err)
	}
	
	config, err := GetKubeConfig()
	
	if err != nil {
		errorLog.Fatal(err)
	}
	
	client, err := NewKubeClient(config)
	
	if err != nil {
		errorLog.Fatal(err)
	}
	
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		cfg:      &cfg,
		client:   client,
	}
	
	tlsPair, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	
	if err != nil {
		errorLog.Fatalln("Error loading TLS certs", err)
	}
	
	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", cfg.Port), // Listen on all the interfaces
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsPair}},
	}
	
	server.Handler = app.setupRoutes()
	
	go func() {
		infoLog.Printf("Starting the web server on port %v", cfg.Port)
		errorLog.Println(server.ListenAndServeTLS("", ""))
	}()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	app.infoLog.Println("Got shutdown signal, shutting down the web server")
	
	if err := server.Shutdown(context.Background()); err != nil {
		errorLog.Fatal("failed to shutdown the web server gracefully", err)
	}
	
}
