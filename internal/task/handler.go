// Copyright 2015 Peter Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Path specifies the task resource path.
const Path = "/task/"

type errRequest struct {
	error
	code int
}

func (e *errRequest) Error() string {
	return fmt.Sprintf("%d %v", e.code, e.error)
}

func badRequestError(err error) *errRequest {
	return &errRequest{err, http.StatusBadRequest}
}

func notFoundError(err error) *errRequest {
	return &errRequest{err, http.StatusNotFound}
}

type handler func(http.ResponseWriter, *http.Request) error

var tasks = NewManager()

var filters = map[string]Filter{
	"isDone":      func(t *Task) bool { return t.Done },
	"isNotDone":   func(t *Task) bool { return !t.Done },
	"isScheduled": func(t *Task) bool { return t.Date != 0 },
}

var sorters = map[string]Sort{
	"dateAsc":      func(t1, t2 *Task) bool { return t1.Date <= t2.Date },
	"dateDesc":     func(t1, t2 *Task) bool { return t1.Date >= t2.Date },
	"priorityAsc":  func(t1, t2 *Task) bool { return t1.Priority <= t2.Priority },
	"priorityDesc": func(t1, t2 *Task) bool { return t1.Priority >= t2.Priority },
}

// RestAPI is a handler function that handles http requests to the task resources.
func RestAPI(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		if len(r.URL.Path) > len(Path) {
			err = read(w, r)
		} else {
			err = readAll(w, r)
		}
	case "POST":
		err = create(w, r)
	case "PUT":
		if len(r.URL.Path) > len(Path) {
			err = update(w, r)
		}
	case "DELETE":
		if len(r.URL.Path) > len(Path) {
			err = delete(w, r)
		}
	default:
		err = badRequestError(fmt.Errorf("%s doesn't implemented", r.Method))
	}
	errorHandler(w, err)
}

// errorHandler handles error responses.
func errorHandler(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	switch e := err.(type) {
	case *errRequest:
		http.Error(w, e.Error(), e.code)
	default:
		log.Println(e)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// create handles requests for the creation of a new task.
func create(w http.ResponseWriter, r *http.Request) error {
	req := struct {
		Title string `json:"title"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return badRequestError(err)
	}
	_, err := tasks.Create(req.Title)
	if err != nil {
		return badRequestError(err)
	}
	return nil
}

// read handles requests for the reads of a specific task.
func read(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequestError(err)
	}
	t, ok := tasks.Find(id)
	if !ok {
		return notFoundError(fmt.Errorf("task id: %d doesn't exists", id))
	}
	return json.NewEncoder(w).Encode(t)
}

// readAll handles requests for the reads of all tasks.
func readAll(w http.ResponseWriter, r *http.Request) error {
	t := tasks.All()

	// Apply filter.
	byFieldEq, ok := filters[r.URL.Query().Get("filter")]
	if ok {
		t = Filter(byFieldEq).Tasks(t)
	}

	// Apply sorter.
	byField, ok := sorters[r.URL.Query().Get("sortBy")]
	if ok {
		Sort(byField).Tasks(t)
	}

	res := struct {
		Tasks []*Task `json:"tasks"`
	}{
		t,
	}
	return json.NewEncoder(w).Encode(res)
}

// update handles requests for the updates of a specific task.
func update(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequestError(err)
	}
	t := new(Task)
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		return badRequestError(err)
	}
	if t.ID != id {
		return badRequestError(fmt.Errorf("inconsistent task IDs"))
	}
	if _, ok := tasks.Find(id); !ok {
		return notFoundError(fmt.Errorf("task id: %d doesn't exists", id))
	}
	return tasks.Update(t)
}

// delete handles requests for the deletion of a specific task.
func delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequestError(err)
	}
	return tasks.Delete(id)
}

// parseID extracts an task id from the request.
func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(r.URL.Path[len(Path):])
}
