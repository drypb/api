package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.checkHealthHandler)

	router.HandlerFunc(http.MethodPost, "/v1/analysis", app.startAnalysisHandler)

	router.HandlerFunc(http.MethodGet, "/v1/report/:id", app.getReportHandler)

	router.HandlerFunc(http.MethodGet, "/v1/status/:id", app.getStatusHandler)

	return app.recoverPanic(router)
}
