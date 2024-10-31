package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.c3sl.ufpr.br/saci/api/internal/analysis"
	"gitlab.c3sl.ufpr.br/saci/api/internal/data"
)

// getReportHandler returns the report of a analysis.
func (app *application) getReportHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		if err == ErrInvalidID || err == ErrEmptyID {
			app.badRequestResponse(w, r, err)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	reportPath := filepath.Join(data.DefaultReportPath, id+".json")
	report, err := os.Open(reportPath)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var data analysis.Report
	decoder := json.NewDecoder(report)
	err = decoder.Decode(&data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"analysis": data}, nil)
}
