package action

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type Task struct {
	Input  interface{}
	Output interface{}
	Err    error
	Func   func(interface{}) (interface{}, error)
}

func ProcessTasks(tasks []Task) error {
	numWorkers := 4
	queue := make(chan Task, len(tasks))
	results := make(chan Task, len(tasks))

	// Enqueue all the tasks onto the queue.
	for _, task := range tasks {
		queue <- task
	}
	close(queue)

	// Start the worker goroutines.
	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range queue {
				output, err := task.Func(task.Input)
				task.Output = output
				task.Err = err
				results <- task
			}
		}()
	}

	// Collect the results from the worker goroutines.
	var err error
	for i := 0; i < len(tasks); i++ {
		task := <-results
		if task.Err != nil {
			err = task.Err
			break
		}
	}
	close(results)

	return err
}

func TestGoroutineQueue(t *testing.T) {
	tasks := []Task{
		{
			Input:  11,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  10,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  9,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  8,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  7,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  6,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  5,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  4,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  3,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
		{
			Input:  2,
			Output: 2,
			Err:    errors.New("error 1"),
			Func: func(i interface{}) (interface{}, error) {
				fmt.Println(fmt.Sprintf("task 1 %v", i))
				time.Sleep(time.Second * time.Duration(i.(int)))
				return i, nil
			},
		},
	}

	if err := ProcessTasks(tasks); err != nil {
		t.Error(err)
	}

}
