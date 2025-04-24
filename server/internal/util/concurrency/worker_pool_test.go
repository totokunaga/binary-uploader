package concurrency

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "BasicFunctionality",
			test: func(t *testing.T) {
				numWorkers := 4
				wp := NewWorkerPool(numWorkers)
				ctx := context.Background()
				wp.Start(ctx)

				var counter int32

				numTasks := 10
				for range make([]struct{}, numTasks) {
					wp.Submit(func() error {
						atomic.AddInt32(&counter, 1)
						return nil
					})
				}
				wp.Wait()

				if atomic.LoadInt32(&counter) != int32(numTasks) {
					t.Errorf("Expected %d tasks to be executed, got %d", numTasks, counter)
				}

				select {
				case err := <-wp.Errors():
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				default:
				}
			},
		},
		{
			name: "ErrorHandling",
			test: func(t *testing.T) {
				numWorkers := 2
				wp := NewWorkerPool(numWorkers)
				ctx := context.Background()
				wp.Start(ctx)

				expectedErr := errors.New("test error")

				wp.Submit(func() error {
					return nil
				})

				wp.Submit(func() error {
					return expectedErr
				})
				wp.Wait()

				var gotErr error
				select {
				case err := <-wp.Errors():
					gotErr = err
				default:
					t.Fatal("Expected an error but got none")
				}

				if gotErr != expectedErr {
					t.Errorf("Expected error %v, got %v", expectedErr, gotErr)
				}
			},
		},
		{
			name: "CancelContext",
			test: func(t *testing.T) {
				numWorkers := 2
				wp := NewWorkerPool(numWorkers)

				ctx, cancel := context.WithCancel(context.Background())
				wp.Start(ctx)

				var wg sync.WaitGroup
				var processedAfterCancel atomic.Int32

				blocker := make(chan struct{})
				cancelled := make(chan struct{})

				for range make([]struct{}, numWorkers) {
					wg.Add(1)
					wp.Submit(func() error {
						defer wg.Done()

						select {
						case <-blocker:
							select {
							case <-cancelled:
								processedAfterCancel.Add(1)
							default:
							}
							return nil
						case <-time.After(50 * time.Millisecond):
							return nil
						}
					})
				}

				cancel()
				close(cancelled)
				time.Sleep(100 * time.Millisecond)

				close(blocker)
				wg.Wait()
				wp.Wait()

				if processedAfterCancel.Load() > 0 {
					t.Errorf("Workers processed %d tasks after context cancellation", processedAfterCancel.Load())
				}
			},
		},
		{
			name: "MultipleErrors",
			test: func(t *testing.T) {
				numWorkers := 3
				wp := NewWorkerPool(numWorkers)
				ctx := context.Background()
				wp.Start(ctx)

				errCount := numWorkers + 2
				for range make([]struct{}, errCount) {
					wp.Submit(func() error {
						return errors.New("test error")
					})
				}
				wp.Wait()

				receivedErrors := 0
				for range wp.Errors() {
					receivedErrors++
				}
				if receivedErrors < cap(wp.errChan) {
					t.Errorf("Expected at least %d errors, got %d", cap(wp.errChan), receivedErrors)
				}
				if receivedErrors > errCount {
					t.Errorf("Got more errors (%d) than tasks with errors (%d)", receivedErrors, errCount)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}
