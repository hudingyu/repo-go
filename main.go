package main

import (
	"errors"
	"fmt"
	"repo-go/actor"
	"repo-go/promise"
	"time"
)

func runPromiseDemo() {
	promise1 := promise.NewPromise(func() (interface{}, error) {
		time.Sleep(time.Second)
		return 1, nil
	})

	promise2 := promise.NewPromise(
		func() (interface{}, error) {
			ch := make(chan int)
			var err error
			go func() {
				time.Sleep(2 * time.Second)
				err = errors.New("手动抛出错误")
				ch <- 1
			}()
			<-ch
			return nil, err
		},
	).Catch(
		func(err error) error {
			fmt.Printf("err: %v\n", err)
			err = errors.New("捕获错误")
			return err
		})

	promise3 := promise.NewPromise(
		func() (interface{}, error) {
			ch := make(chan int)
			go func() {
				time.Sleep(1 * time.Second)
				ch <- 3
			}()
			v := <-ch
			return v, nil
		},
	).ThenSuccess(
		func(v interface{}) (interface{}, error) {
			return v.(int) + 1, nil
		},
	)

	promiseAll1 := promise.All(promise1, promise2)
	value1, err := promiseAll1.Done()
	if err != nil {
		fmt.Printf("err1: %v\n", err)
	} else {
		fmt.Printf("value1: %+v\n", value1)
	}

	promiseAll2 := promise.All(promise1, promise3)
	value2, err := promiseAll2.Done()
	if err != nil {
		fmt.Printf("err2: %v\n", err)
	} else {
		fmt.Printf("value2: %+v\n", value2)
	}
}

func runActorDemo() {
	actor := actor.NewActor(10)
	defer actor.Close()

	task := func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 10, nil
	}

	value, _ := actor.Do(task).ThenSuccess(
		func(v interface{}) (interface{}, error) {
			return v.(int) + 1, nil
		},
	).Done()

	fmt.Printf("value1: %+v\n", value)
}

func main() {
	// runPromiseDemo()
	runActorDemo()
}
