package scheduler

import "sync"

type TaskQueue struct {
	ch      chan Task
	mu      sync.RWMutex
	stopped bool
	closed  bool
}

func NewTaskQueue(bufferSize int) *TaskQueue {
	return &TaskQueue{
		ch: make(chan Task, bufferSize),
	}
}

func (q *TaskQueue) AddTask(t Task) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.stopped {
		return false
	}
	select {
	case q.ch <- t:
		return true
	default:
		return false
	}
}

func (q *TaskQueue) Chan() <-chan Task {
	return q.ch
}

func (q *TaskQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stopped = true
}

func (q *TaskQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	close(q.ch)
	q.closed = true
}
