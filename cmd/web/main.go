package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	/*
		ENV_SERVER_LISTEN, configures where the server should listen. E.g. use :80 to listen on port 80 on all interfaces
	*/
	ENV_SERVER_LISTEN = "SERVER_LISTEN"
	/*
		ENV_USE_SSL, control whether ssl should be used. If set to "true" we expect a "cert.pem" and a "privkey.pem"
		file in the "private" subfolder
	*/
	ENV_USE_SSL          = "USE_SSL"
	ENV_SSL_CERT_PATH    = "SSL_CERT_PATH"
	ENV_SSL_PRIVKEY_PATH = "SSL_PRIVKEY_PATH"
)

type application struct {
	//Concurrency safe
	errorLog *log.Logger
	//Concurrency safe
	infoLog *log.Logger
	//Run jobs at specific times ar after a certain amount of time
	scheduler *gocron.Scheduler
	//Data Model for served Dishes
	menuModel *MenuCache
}

func main() {

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	addr := os.Getenv(ENV_SERVER_LISTEN)
	if addr == "" {
		addr = ":8080"
	}

	mc, err := NewMenuCache(errorLog, infoLog)
	if err != nil {
		errorLog.Fatalf("NewMenuCache: %v", err)
	}

	app := &application{
		errorLog:  errorLog,
		infoLog:   infoLog,
		scheduler: gocron.NewScheduler(time.Local),
		menuModel: mc,
	}

	//refresh daily to reduce risk of long query due to refresh
	_, err = app.scheduler.Every(1).Day().At("01:00").Do(func() {
		err := app.menuModel.Refresh()
		if err != nil {
			app.errorLog.Printf("Peridic menuModel.Refresh call failed\n")
		}
	})
	if err != nil {
		app.errorLog.Fatalf("Failed to scheule refresh job: %v", err)
	}
	app.scheduler.StartAsync()
	app.infoLog.Printf("Registered periodic call to menuModel.Refresh\n")

	// Set the server's TLSConfig field to use the tlsConfig variable we just
	// created.
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
		ErrorLog:     errorLog,
	}
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	signal.Notify(terminate, os.Kill)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-terminate
		app.infoLog.Printf("Got terminate signal")
		shutdownCtx, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
		if err := srv.Shutdown(shutdownCtx); err != nil {
			app.errorLog.Printf("Clean shutdown failed: %v\n", err)
		} else {
			app.infoLog.Printf("Clean Shutdown done\n")
		}
	}()

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		app.errorLog.Fatalf("Failed to listen on tcp4 %v: %v\n", addr, err)
	}
	infoLog.Printf("Starting server on %s", addr)
	if os.Getenv(ENV_USE_SSL) == "true" {
		err = srv.ServeTLS(listener, os.Getenv(ENV_SSL_CERT_PATH), os.Getenv(ENV_SSL_PRIVKEY_PATH))
	} else {
		err = srv.Serve(listener)
	}
	if err != nil && err != http.ErrServerClosed {
		errorLog.Fatal(err)
	}
	wg.Wait()
}
