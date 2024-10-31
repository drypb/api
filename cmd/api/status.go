package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/drypb/api/internal/data"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 1 * time.Second
)

func ReadFileIfModified(filePath string, lastMod time.Time) ([]byte, time.Time, error) {
	fi, err := os.Stat(filePath)

	if err != nil {
		return nil, lastMod, err
	}

	if !fi.ModTime().After(lastMod) {
		return nil, lastMod, nil
	}

	p, err := os.ReadFile(filePath)

	if err != nil {
		return nil, fi.ModTime(), err
	}

	return p, fi.ModTime(), nil
}

func Reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func Writer(ws *websocket.Conn, filePath string, lastMod time.Time) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-fileTicker.C:
			var p []byte
			var err error

			p, lastMod, err = ReadFileIfModified(filePath, lastMod)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			if p != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (app *application) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		if err == ErrInvalidID || err == ErrEmptyID {
			app.badRequestResponse(w, r, err)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		_, ok := err.(websocket.HandshakeError)
		if !ok {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// optionally, the initial value of lastMod can be set as a query parameter
	lastMod := time.Time{}
	n, err := strconv.ParseInt(r.FormValue("lastMod"), 16, 64)
	if err == nil {
		lastMod = time.Unix(0, n)
	}

	path := filepath.Join(data.DefaultStatusPath, id+".json")

	go Writer(ws, path, lastMod)
	Reader(ws)
}
