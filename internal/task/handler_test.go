// Copyright 2015 Peter Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func checkStatusCode(got, want int) error {
	if got != want {
		return fmt.Errorf("got status code %d; want %d", got, want)
	}
	return nil
}

func TestErrorHandlerInternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	errorHandler(rec, errors.New("Internal Error"))
	if err := checkStatusCode(rec.Code, http.StatusInternalServerError); err != nil {
		t.Error(err)
	}
}

func TestCreateReq(t *testing.T) {
	tasks = NewManager()
	for _, test := range []struct {
		json string
		code int
	}{
		{`{"title":"new"}`, http.StatusOK},
		{`{"title":"}`, http.StatusBadRequest},
		{`{}`, http.StatusBadRequest},
	} {
		req, err := http.NewRequest("POST", Path, bytes.NewBufferString(test.json))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		RestAPI(rec, req)

		if err := checkStatusCode(rec.Code, test.code); err != nil {
			t.Errorf("HTTP request %v: %v", req, err)
			t.Errorf("Request body: %v", test.json)
			t.Errorf("Recieve body: %q", rec.Body)
		}
	}
}

func TestReadReq(t *testing.T) {
	tasks = NewManager()
	for _, title := range []string{
		"New Task 0",
		"New Task 1",
		"New Task 2",
	} {
		task, err := tasks.Create(title)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("GET", Path+strconv.Itoa(task.ID), nil)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		RestAPI(rec, req)

		if err := checkStatusCode(rec.Code, http.StatusOK); err != nil {
			t.Errorf("HTTP request %v: %v", req, err)
			t.Errorf("Recieve body: %q", rec.Body)
		}

		body := rec.Body.String()
		json, err := json.Marshal(task)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := body[:len(body)-1], string(json); got != want {
			t.Errorf("HTTP request %v\n got %q\nwant %q", req, got, want)
		}
	}
}

func TestReadReqError(t *testing.T) {
	for _, test := range []struct {
		id   string
		code int
	}{
		{"-1", http.StatusNotFound},
		{"wrongID", http.StatusBadRequest},
	} {
		req, err := http.NewRequest("GET", Path+test.id, nil)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		RestAPI(rec, req)

		if err := checkStatusCode(rec.Code, test.code); err != nil {
			t.Errorf("HTTP request %v: %v", req, err)
			t.Errorf("Recieve body: %q", rec.Body)
		}
	}
}

func TestReadAllReq(t *testing.T) {
	tasks = NewManager()
	var tt []*Task
	for _, title := range []string{
		"New Task 0",
		"New Task 1",
		"New Task 2",
	} {
		task, err := tasks.Create(title)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("GET", Path, nil)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		RestAPI(rec, req)

		if err := checkStatusCode(rec.Code, http.StatusOK); err != nil {
			t.Errorf("HTTP request %v: %v", req, err)
			t.Errorf("Recieve body: %q", rec.Body)
		}

		tt = append(tt, task)
		body := rec.Body.String()
		json, err := json.Marshal(
			struct {
				Tasks []*Task `json:"tasks"`
			}{
				tt,
			})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := body[:len(body)-1], string(json); got != want {
			t.Errorf("HTTP request %v\n got %q\nwant %q", req, got, want)
		}
	}
}

func TestUpdateReq(t *testing.T) {
	tasks = NewManager()

	want, err := tasks.Create("New Task")
	if err != nil {
		t.Fatal(err)
	}
	want.Title = "Updated Task"

	json := `{"Title":"` + want.Title + `"}`
	req, err := http.NewRequest("PUT", Path+strconv.Itoa(want.ID), bytes.NewBufferString(json))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	RestAPI(rec, req)

	if err := checkStatusCode(rec.Code, http.StatusOK); err != nil {
		t.Errorf("HTTP request %v: %v", req, err)
		t.Errorf("Request body: %v", json)
		t.Errorf("Recieve body: %q", rec.Body)
	}

	got, ok := tasks.Find(want.ID)
	if !ok {
		t.Fatalf("Find: task with id %d doesn't exist", want.ID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("HTTP request %v\n got %v\nwant %v", req, got, want)
	}
}

func TestUpdateReqError(t *testing.T) {
	tasks = NewManager()
	for _, test := range []struct {
		id   string
		json string
		code int
	}{
		{"0", `{"id":}`, http.StatusBadRequest},
		{"0", `{"id":0,"title":"New Title"}`, http.StatusNotFound},
		{"-1", `{"id":0,"title":"New Title"}`, http.StatusBadRequest},
		{"wrongID", `{"id":0,"title":"New Title"}`, http.StatusBadRequest},
	} {
		req, err := http.NewRequest("PUT", Path+test.id, bytes.NewBufferString(test.json))
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		RestAPI(rec, req)

		if err := checkStatusCode(rec.Code, test.code); err != nil {
			t.Errorf("HTTP request %v: %v", req, err)
			t.Errorf("Request body: %v", test.json)
			t.Errorf("Recieve body: %q", rec.Body)
		}
	}
}

func TestDeleteReq(t *testing.T) {
	tasks = NewManager()
	task, err := tasks.Create("New Task")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("DELETE", Path+strconv.Itoa(task.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	RestAPI(rec, req)

	if err := checkStatusCode(rec.Code, http.StatusOK); err != nil {
		t.Errorf("HTTP request %v: %v", req, err)
		t.Errorf("Recieve body: %q", rec.Body)
	}

	if tasks.Count() != 0 {
		t.Errorf("HTTP request %v: failed to delete task: %v", req, task)
	}
}

func TestDeleteReqError(t *testing.T) {
	req, err := http.NewRequest("DELETE", Path+"wrongID", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	RestAPI(rec, req)

	if err := checkStatusCode(rec.Code, http.StatusBadRequest); err != nil {
		t.Errorf("HTTP request %v: %v", req, err)
		t.Errorf("Recieve body: %q", rec.Body)
	}
}

func TestUnknownReq(t *testing.T) {
	req, err := http.NewRequest("UNKNOWN", Path, nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	RestAPI(rec, req)

	if err := checkStatusCode(rec.Code, http.StatusBadRequest); err != nil {
		t.Errorf("HTTP request %v: %v", req, err)
		t.Errorf("Recieve body: %q", rec.Body)
	}
}
