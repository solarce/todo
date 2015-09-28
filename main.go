// Copyright 2015 Peter Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"

	"github.com/mrekucci/todo/internal/task"
)

// corsHeaders wraps http.HandlerFunc and adds CORS headers to response.
// Function also handles pre-flight requests.
func corsHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" { // Stop the pre-flight request.
			return
		}
		fn(w, r)
	}
}

func main() {
	http.Handle(task.Path, http.HandlerFunc(corsHeaders(task.RestAPI)))
	http.Handle("/", http.FileServer(http.Dir("frontend/web")))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
