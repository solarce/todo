// Copyright 2015 Peter Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"errors"
	"sort"
)

// ErrCreateEmptyTitle indicates attempt to create task with an empty title.
var ErrCreateEmptyTitle = errors.New("Create: empty title")

// ErrUpdateUnknown indicates attempt to update unknown task.
var ErrUpdateUnknown = errors.New("Update: unknown task")

// ErrDeleteUnknown indicates attempt to delete unknown task.
var ErrDeleteUnknown = errors.New("Delete: unknown task")

// Task enumerates task properties.
type Task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Date     int64  `json:"date"`
	Note     string `json:"note"`
	Priority byte   `json:"priority"`
	Done     bool   `json:"done"`
}

// Sort is the type of a Sort.Less function that
// defines the ordering of its Task arguments.
type Sort func(t1, t2 *Task) bool

// Tasks sorts the argument slice according to the Sort function.
func (s Sort) Tasks(tasks []*Task) {
	ts := &sorter{
		tasks: tasks,
		by:    s,
	}
	sort.Sort(ts)
}

// sorter joins a By function and a slice of tasks to be sorted.
type sorter struct {
	tasks []*Task
	by    Sort
}

// Len is part of sort.Interface.
func (s *sorter) Len() int { return len(s.tasks) }

// Swap is part of sort.Interface.
func (s *sorter) Swap(i, j int) { s.tasks[i], s.tasks[j] = s.tasks[j], s.tasks[i] }

// Less is part of sort.Interface.
// It calls the "by" closure in the sorter.
func (s *sorter) Less(i, j int) bool { return s.by(s.tasks[i], s.tasks[j]) }

// Filter is function that defines the filtering of its Task arguments.
type Filter func(t *Task) bool

// Tasks filters the argument slice according to the Filter function.
func (f Filter) Tasks(tasks []*Task) []*Task {
	var r []*Task
	for _, t := range tasks {
		if f(t) {
			r = append(r, t)
		}
	}
	return r
}

// Manager defines operation of task storage.
type Manager interface {
	// Returns new task with given title.
	// An error is returned if task was not created successfully.
	Create(title string) (*Task, error)

	// Returns task with given id.
	// Empty Task and false is returned if a task with such an id doesn't exist.
	Find(id int) (task *Task, ok bool)

	// Returns all stored tasks.
	All() []*Task

	// Updates given task.
	// An error is returned if such a task doesn't exist.
	Update(task *Task) error

	// Deletes task with given id.
	// An error is returned if a task with such id doesn't exist.
	Delete(id int) error

	// Returns a number of stored tasks.
	Count() int
}

// NewManager returns a new empty Manager.
func NewManager() Manager {
	return &inMemory{}
}

// inMemory allows manage tasks in memory.
type inMemory struct {
	tasks  []*Task
	nextID int
}

// Create stores and returns new task with given title.
// An error is returned if the title is empty.
func (m *inMemory) Create(title string) (*Task, error) {
	if title == "" {
		return nil, ErrCreateEmptyTitle
	}
	t := &Task{ID: m.nextID, Title: title}
	m.tasks = append(m.tasks, t)
	m.nextID++
	return t, nil
}

// Find returns task with given id.
// Returns empty Task and false, if a task with such id doesn't exist.
func (m *inMemory) Find(id int) (task *Task, ok bool) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, true
		}
	}
	return nil, false
}

// All returns all stored tasks.
func (m *inMemory) All() []*Task {
	return m.tasks
}

// Update updates given task.
// Returns error if such a task doesn't exist.
func (m *inMemory) Update(task *Task) error {
	for i, t := range m.tasks {
		if t.ID == task.ID {
			c := *task // Copy the task to save the changes.
			m.tasks[i] = &c
			return nil
		}
	}
	return ErrUpdateUnknown
}

// Delete deletes task with given id.
// Returns an error if a task with such id doesn't exist.
func (m *inMemory) Delete(id int) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks, m.tasks[m.Count()-1] = append(m.tasks[:i], m.tasks[i+1:]...), nil
			return nil
		}
	}
	return ErrDeleteUnknown
}

// Count returns a number of stored tasks.
func (m *inMemory) Count() int {
	return len(m.tasks)
}
