package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

/*
aliveHandler just writs "I am aliveHandler"  to w and can be used by external agents
to test if the server is running
*/
func (app *application) aliveHandler(w http.ResponseWriter, _ *http.Request) {
	if _, err := fmt.Fprintf(w, "I am aliveHandler\n"); err != nil {
		app.errorLog.Printf("failed to send aliveHandler message: %v\n", err)
	}
}

/*
menuHandler returns all dishes for the day specified in the url as json
*/
func (app *application) menuHandler(w http.ResponseWriter, r *http.Request) {

	date, err := time.ParseInLocation("2006-01-02", r.URL.Query().Get(":date"), time.Local)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Pass date as yyyy-mm-dd\n")); err != nil {
			app.errorLog.Printf("Failed to write error response\n")
		}

		return
	}

	dishes, err := app.menuModel.GetMenu(date)
	if err != nil {
		if errors.Is(err, invDateError) {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Date either to far in the past or to far in the future\n")); err != nil {
				app.errorLog.Printf("Failed to write error response\n")
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(dishes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, bytes.NewReader(response)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return

}
