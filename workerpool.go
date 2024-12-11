package workerpool

import "sync"

// Response represents a response from a worker.
//
// Result - string, slice, or another type, which returns a task (function).
// Err - error returned from a task.
type Response struct {
	Result interface{}
	Err    error
}

// Task contains properties for a worker pool.
// Fn - function to execute tasks.
// Resp - channel to send response.
type Task struct {
	Fn   func() (interface{}, error)
	Resp chan<- Response
}

// WorkerPool manages a pool of workers.
type WorkerPool struct {
	tasks chan Task
	wg    *sync.WaitGroup
}

// New creates a new pool of workers.
func New(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		tasks: make(chan Task),
		wg:    new(sync.WaitGroup),
	}

	pool.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go pool.worker()
	}

	return pool
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for task := range wp.tasks {
		result, err := task.Fn()
		task.Resp <- Response{
			Result: result,
			Err:    err,
		}
	}
}

// Submit adds a task to the pool.
func (wp *WorkerPool) Submit(fn func() (interface{}, error)) chan Response {
	resp := make(chan Response, 1)
	wp.tasks <- Task{Fn: fn, Resp: resp}

	return resp
}

// Wait shuts down the pool.
func (wp *WorkerPool) Wait() {
	close(wp.tasks)
	wp.wg.Wait()
}
