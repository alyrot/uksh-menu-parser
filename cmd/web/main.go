package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	ENV_SERVER_LISTEN = "SERVER_LISTEN"
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
	go func() {
		<-terminate
		srv.Shutdown(context.TODO())
	}()

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		app.errorLog.Fatalf("Failed to listen on tcp4 %v: %v\n", addr, err)
	}
	infoLog.Printf("Starting server on %s", addr)
	//err = srv.ServeTLS(listener, "../../private/ssl/cert.pem", "../../private/ssl/privkey.pem")
	err = srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		errorLog.Fatal(err)
	}
}
