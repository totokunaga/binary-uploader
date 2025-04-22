package concurrency

import (
	"context"
	"sync"
)

// Task represents a function that can be executed by a worker.
type Task func() error

// WorkerPool manages a pool of workers that process tasks concurrently.
type WorkerPool struct {
	tasks   chan Task
	wg      sync.WaitGroup
	errChan chan error
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		tasks:   make(chan Task),
		errChan: make(chan error, numWorkers),
	}
}

// Start initializes and starts the worker goroutines.
func (wp *WorkerPool) Start(ctx context.Context) {
	for range make([]struct{}, cap(wp.errChan)) {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for {
				select {
				case task, ok := <-wp.tasks:
					if !ok {
						return
					}
					if err := task(); err != nil {
						select {
						case wp.errChan <- err:
						default:
							// Buffer full, errors might be lost
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Submit adds a task to the worker pool for processing.
func (wp *WorkerPool) Submit(task Task) {
	wp.tasks <- task
}

// Wait closes the task channel and waits for all workers to finish.
func (wp *WorkerPool) Wait() {
	close(wp.tasks)
	wp.wg.Wait()
	close(wp.errChan)
}

// Errors returns a channel that provides any errors from task execution.
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errChan
}
