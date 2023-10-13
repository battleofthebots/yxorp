package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/google/shlex"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

/*
Stolen from https://github.com/elithrar/admission-control/blob/master/request_logger.go
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
	return
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("[!] %s: %s", err, debug.Stack())
			}
		}()

		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s - %d", r.Method, r.URL.EscapedPath(), wrapped.status)
	}

	return http.HandlerFunc(fn)
}

func requireInternal(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isInternal := strings.HasPrefix(r.RemoteAddr, "127.0.0.2:")
		if isInternal {
			h.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
}

func debugEndpoint(w http.ResponseWriter, r *http.Request) {
	output, err := run(r.URL.Query().Get("cmd"))

	var e2 string
	if err != nil {
		e2 = err.Error()
	}
	res := map[string]interface{}{
		"Method":        r.Method,
		"Host":          r.Host,
		"RemoteAddr":    r.RemoteAddr,
		"Headers":       r.Header,
		"Error":         e2,
		"CommandOutput": output,
	}

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	if err := e.Encode(res); err != nil {
		log.Println(err)
	}
}

func run(cmd string) (string, error) {
	if strings.TrimSpace(cmd) == "" {
		return "", nil
	}
	args, err := shlex.Split(cmd)
	if err != nil {
		return "", err
	}
	cmd = args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	out := new(bytes.Buffer)
	c := exec.Command(cmd, args...)
	c.Stderr = out
	c.Stdout = out
	err = c.Run()
	return out.String(), err
}

func main() {
	port := 80
	log.Printf("listening on :%d", port)
	a := mux.NewRouter()
	a.Use(LoggingMiddleware)
	a.Use(handlers.ProxyHeaders)
	dbg := a.Host("localhost.localdomain").Subrouter()
	dbg.Use(requireInternal)
	dbg.HandleFunc("/debug", debugEndpoint)
	a.PathPrefix("/").Handler(http.FileServer(http.Dir("")))
	http.ListenAndServe(fmt.Sprintf(":%d", port), a)
}
