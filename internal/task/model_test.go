// Copyright 2015 Peter Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"reflect"
	"testing"
)

var unsortedTestTasks = [...]Task{
	Task{Title: "Task 0", Priority: 1, Done: false},
	Task{Title: "Task 1", Date: 1426691590, Priority: 2, Done: true},
	Task{Title: "Task 3", Date: 1426691592, Priority: 0, Done: false},
	Task{Title: "Task 2", Date: 1426691591, Priority: 1, Done: true},
}

var testTasks = [...]Task{
	Task{ID: 0, Title: "Task 0"},
	Task{ID: 1, Title: "Task 1"},
	Task{ID: 2, Title: "Task 2"},
}

// addTasks populates m with tasks.
func addTasks(m Manager, tasks []Task, t *testing.T) error {
	for _, task := range tasks {
		if _, err := m.Create(task.Title); err != nil {
			return err
		}
	}
	for _, task := range tasks {
		if err := m.Update(&task); err != nil {
			return err
		}
	}
	return nil
}

// valToPtr takes []Task and returns a []*Task.
func valToPtr(tasks []Task) []*Task {
	var r []*Task
	for _, t := range tasks {
		c := t // Copy the task to store the pointer to it.
		r = append(r, &c)
	}
	return r
}

// ptrToVal takes []*Task and returns a []Task.
func ptrToVal(tasks []*Task) []Task {
	var r []Task
	for _, t := range tasks {
		r = append(r, *t)
	}
	return r
}

func TestSort(t *testing.T) {
	for _, test := range []struct {
		name    string
		byField Sort
	}{
		{"increasing date", func(t1, t2 *Task) bool { return t1.Date <= t2.Date }},
		{"decreasing date", func(t1, t2 *Task) bool { return t1.Date >= t2.Date }},
		{"increasing priority:", func(t1, t2 *Task) bool { return t1.Priority <= t2.Priority }},
		{"decreasing priority:", func(t1, t2 *Task) bool { return t1.Priority >= t2.Priority }},
	} {
		data := valToPtr(unsortedTestTasks[:])
		Sort(test.byField).Tasks(data)
		for i := 0; i < len(data)-1; i++ {
			if !test.byField(data[i], data[i+1]) {
				t.Errorf("Sort by %q\n in %v\ngot %v", test.name, unsortedTestTasks, ptrToVal(data))
			}
		}
	}
}

func TestFilter(t *testing.T) {
	for _, test := range []struct {
		name      string
		byFieldEq Filter
	}{
		{"Done == True", func(t *Task) bool { return t.Done }},
		{"Done == False", func(t *Task) bool { return !t.Done }},
		{"Date != 0", func(t *Task) bool { return t.Date != 0 }},
	} {
		data := valToPtr(unsortedTestTasks[:])
		var want []*Task
		for _, task := range data {
			if test.byFieldEq(task) {
				want = append(want, task)
			}
		}
		got := Filter(test.byFieldEq).Tasks(data)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Filter by %q\n  in %v\n got %v\nwant %v", test.name, unsortedTestTasks, ptrToVal(got), ptrToVal(want))
		}
	}
}

func TestCreate(t *testing.T) {
	m := NewManager()
	for _, test := range []struct {
		in   string
		want *Task
		err  error
	}{
		{testTasks[0].Title, &testTasks[0], nil},
		{testTasks[1].Title, &testTasks[1], nil},
		{testTasks[2].Title, &testTasks[2], nil},
		{"", nil, ErrCreateEmptyTitle},
	} {
		got, err := m.Create(test.in)
		if !reflect.DeepEqual(got, test.want) || err != test.err {
			t.Errorf("Create(%q) = %v, %v; want %v, %v", test.in, got, err, test.want, test.err)
		}
	}
}

func TestFind(t *testing.T) {
	m := NewManager()
	addTasks(m, testTasks[:], t)
	for _, test := range []struct {
		in   int
		want *Task
		ok   bool
	}{
		{testTasks[0].ID, &testTasks[0], true},
		{testTasks[1].ID, &testTasks[1], true},
		{testTasks[2].ID, &testTasks[2], true},
		{len(testTasks), nil, false},
	} {
		if got, ok := m.Find(test.in); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Find(%d) = %v, %t; want %v, %t", test.in, got, ok, test.want, test.ok)
		}
	}
}

func TestAll(t *testing.T) {
	m := NewManager()
	data := testTasks[:]
	if err := addTasks(m, data, t); err != nil {
		t.Fatalf("cannot initialize test with tasks due to: %v", err)
	}
	if got, want := m.All(), valToPtr(data); !reflect.DeepEqual(got, want) {
		t.Errorf("All() = %v\n              want %v", ptrToVal(got), ptrToVal(want))

	}
}

func TestUpdate(t *testing.T) {
	m := NewManager()

	// Test update unknown.
	task := Task{}
	if got, want := m.Update(&task), ErrUpdateUnknown; got != want {
		t.Errorf("Update(%v) = %v; want %v", task, got, want)
	}

	// Test update existing.
	if err := addTasks(m, testTasks[:], t); err != nil {
		t.Fatalf("cannot initialize test with tasks due to: %v", err)
	}
	for _, task := range testTasks {
		want := &Task{ID: task.ID, Title: "Updated Title", Note: "Updated Note", Priority: 1, Done: true}
		task.Title = want.Title
		task.Note = want.Note
		task.Priority = want.Priority
		task.Done = want.Done
		if err := m.Update(&task); err != nil {
			t.Errorf("Update(%v): unexpected error: %v", task, err)
		}
		switch got, ok := m.Find(task.ID); {
		case !ok:
			t.Fatalf("Find(%d): failed to find a task: %v", task.ID, task)
		case !reflect.DeepEqual(got, want):
			t.Errorf("Update(%+v)\n got %+v\nwant %+v", task, got, want)

		}
	}
}

func TestDelete(t *testing.T) {
	m := NewManager()

	// Test delete unknown.
	if got, want := m.Delete(0), ErrDeleteUnknown; got != want {
		t.Errorf("Delete(0) = %v; want %v", got, want)
	}

	// Test delete existing.
	if err := addTasks(m, testTasks[:], t); err != nil {
		t.Fatalf("cannot initialize test with tasks due to: %v", err)
	}
	tl := m.Count()
	for _, task := range testTasks {
		if err := m.Delete(task.ID); err != nil {
			t.Errorf("Delete(%d): unexpected error: %v", task.ID, err)
		}
		if tl--; tl != m.Count() {
			t.Errorf("Delete(%d): failed to delete task: %v", task.ID, task)
		}
	}
}

func TestCount(t *testing.T) {
	m := NewManager()
	if err := addTasks(m, testTasks[:], t); err != nil {
		t.Fatalf("cannot initialize test with tasks due to: %v", err)
	}
	if got, want := m.Count(), len(testTasks); got != want {
		t.Errorf("Count() = %d; want %d", got, want)
	}
}
