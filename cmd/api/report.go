package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/drypb/api/internal/analysis"
	"github.com/drypb/api/internal/data"
)

// GetReportHandler returns the report of an analysis.
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
