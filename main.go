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

	promiseAll1 := promise.All(promise1, promise2)
	value1, err1 := promiseAll1.Done()
	if err1 != nil {
		fmt.Printf("err1: %v\n", err1)
	} else {
		fmt.Printf("value1: %+v\n", value1)
	}

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

	promise4 := promise.NewPromise(func() (interface{}, error) {
		time.Sleep(time.Second)
		return 1, nil
	})

	promiseAll2 := promise.All(promise3, promise4)
	value2, err2 := promiseAll2.Done()
	if err2 != nil {
		fmt.Printf("err2: %v\n", err2)
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
