package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	standardMiddleware := alice.New(app.logRequest)
	//supports semantic urls, put exact matches before wildcard matches
	mux := pat.New()
	mux.Get("/alive", http.HandlerFunc(app.aliveHandler))
	mux.Get("/menu/:date", http.HandlerFunc(app.menuHandler))

	return standardMiddleware.Then(mux)
}
