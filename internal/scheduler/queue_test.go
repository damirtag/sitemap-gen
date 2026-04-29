package scheduler

import (
	"testing"
)

func TestAddTask_ReturnsTrue(t *testing.T) {
	q := NewTaskQueue(10)
	ok := q.AddTask(Task{URL: "https://example.com", Depth: 0})
	if !ok {
		t.Error("expected true when queue has space")
	}
}

func TestAddTask_AfterStop(t *testing.T) {
	q := NewTaskQueue(10)
	q.Stop()
	ok := q.AddTask(Task{URL: "https://example.com", Depth: 0})
	if ok {
		t.Error("expected false after Stop()")
	}
}

func TestAddTask_FullQueue(t *testing.T) {
	q := NewTaskQueue(1)
	q.AddTask(Task{URL: "https://a.com", Depth: 0})
	ok := q.AddTask(Task{URL: "https://b.com", Depth: 0})
	if ok {
		t.Error("expected false when queue is full")
	}
}

func TestClose_DrainRemaining(t *testing.T) {
	q := NewTaskQueue(10)
	q.AddTask(Task{URL: "https://a.com", Depth: 0})
	q.AddTask(Task{URL: "https://b.com", Depth: 0})
	q.Close()

	count := 0
	for range q.Chan() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 tasks drained, got %d", count)
	}
}
