package workerpool

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type args struct {
	numWorkers int
	tasks      []func() (interface{}, error)
}

type mockBehaviour func(*WorkerPool, []func() (interface{}, error)) chan chan Response

func TestWorkerPool(t *testing.T) {
	tests := []struct {
		name          string
		args          args
		mockBehaviour mockBehaviour
		expected      []Response
	}{
		{
			name: "SuccessfulTasks",
			args: args{
				numWorkers: 2,
				tasks: []func() (interface{}, error){
					func() (interface{}, error) { return "result1", nil },
					func() (interface{}, error) { return "result2", nil },
				},
			},
			mockBehaviour: func(wp *WorkerPool, tasks []func() (interface{}, error)) chan chan Response {
				responses := make(chan chan Response, len(tasks))
				for _, task := range tasks {
					responses <- wp.Submit(task)
				}

				wp.Wait()
				close(responses)

				return responses
			},
			expected: []Response{
				{Result: "result1", Err: nil},
				{Result: "result2", Err: nil},
			},
		},
		{
			name: "TaskWithError",
			args: args{
				numWorkers: 1,
				tasks: []func() (interface{}, error){
					func() (interface{}, error) { return nil, errors.New("task error") },
				},
			},
			mockBehaviour: func(wp *WorkerPool, tasks []func() (interface{}, error)) chan chan Response {
				responses := make(chan chan Response, len(tasks))
				for _, task := range tasks {
					responses <- wp.Submit(task)
				}

				wp.Wait()
				close(responses)

				return responses
			},
			expected: []Response{
				{Result: nil, Err: errors.New("task error")},
			},
		},
		{
			name: "MultipleTasksWithMixedResults",
			args: args{
				numWorkers: 3,
				tasks: []func() (interface{}, error){
					func() (interface{}, error) { return "result1", nil },
					func() (interface{}, error) { return nil, errors.New("error2") },
					func() (interface{}, error) { return "result3", nil },
				},
			},
			mockBehaviour: func(wp *WorkerPool, tasks []func() (interface{}, error)) chan chan Response {
				responses := make(chan chan Response, len(tasks))
				for _, task := range tasks {
					responses <- wp.Submit(task)
				}

				wp.Wait()
				close(responses)

				return responses
			},
			expected: []Response{
				{Result: "result1", Err: nil},
				{Result: nil, Err: errors.New("error2")},
				{Result: "result3", Err: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := New(tt.args.numWorkers)

			var (
				responses = tt.mockBehaviour(wp, tt.args.tasks)
				i         = 0
			)

			for response := range responses {
				assert.Equal(t, tt.expected[i], <-response)
				i++
			}
		})
	}
}
