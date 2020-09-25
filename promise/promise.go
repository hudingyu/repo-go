package promise

import (
	"sync/atomic"
)

type State string

const (
	PENDING  State = "pending"
	RESOLVED State = "resolved"
	REJECTED State = "rejected"
)

type Promise struct {
	currentState State
	done         chan struct{}
	value        interface{}
	err          error
}

type TaskFunc func() (interface{}, error)
type OnResolved func(value interface{}) (interface{}, error)
type OnRejected func(err error) error

func NewPromise(task TaskFunc) *Promise {
	// promise初始状态为pending
	promise := &Promise{currentState: PENDING, done: make(chan struct{})}

	go func() {
		// 任务完成之后将promise标志为已完成
		defer close(promise.done)
		value, err := task()
		if err != nil {
			//promise rejected后, 保存reason
			promise.currentState = REJECTED
			promise.err = err
		} else {
			//promise fulfilled后, 保存value
			promise.currentState = RESOLVED
			promise.value = value
		}
	}()

	return promise
}

// 实现await的功能
func (p *Promise) Done() (interface{}, error) {
	<-p.done
	return p.value, p.err
}

// then方法传入onFulfilled 和 onRejected方法，作为参数，返回值是一个promise
func (p *Promise) Then(resolveFunc OnResolved, rejectFunc OnRejected) *Promise {
	newPromise := &Promise{currentState: PENDING, done: make(chan struct{})}

	go func() {
		defer close(newPromise.done)
		// 等待当前promise完成
		<-p.done

		if p.currentState == RESOLVED {
			if resolveFunc == nil {
				newPromise.currentState = RESOLVED
				newPromise.value = p.value
				return
			}
			// resolveFunc接收当前promise的value作为参数
			value, err := resolveFunc(p.value)
			if err != nil {
				newPromise.currentState = REJECTED
				newPromise.err = err
			} else {
				newPromise.currentState = RESOLVED
				newPromise.value = value
			}
		} else {
			if rejectFunc == nil {
				newPromise.err = p.err
			} else {
				// rejectFunc接收当前promise的err作为参数
				err := rejectFunc(p.err)
				newPromise.err = err
			}
			newPromise.currentState = REJECTED
		}
	}()

	return newPromise
}

// 只传入OnResolved参数
func (p *Promise) ThenSuccess(resolveFunc OnResolved) *Promise {
	return p.Then(resolveFunc, nil)
}

// 只传入OnRejected参数
func (p *Promise) Catch(rejectFunc OnRejected) *Promise {
	return p.Then(nil, rejectFunc)
}

// 传入一组Promise，全部完成才算完成，任何一个失败都会立即返回错误
func All(promises ...*Promise) *Promise {
	newPromise := &Promise{
		currentState: PENDING,
		done:         make(chan struct{}),
	}

	go func() {
		defer close(newPromise.done)

		errChan := make(chan error)
		valueChan := make(chan interface{})

		values := make([]interface{}, len(promises))
		count := atomic.Value{}
		count.Store(0)

		for index, promise := range promises {
			go func(index int, promise *Promise) {
				v, err := promise.Done()
				// 某个promise出错
				if err != nil {
					errChan <- err
				}
				values[index] = v
				count.Store(count.Load().(int) + 1)
				// 所有promise都成功
				if count.Load().(int) == len(promises) {
					valueChan <- values
				}
			}(index, promise)
		}

		for {
			select {
			//	一旦某个promise出错，立即返回REJECTED
			case err := <-errChan:
				newPromise.err = err
				newPromise.currentState = REJECTED
				return
			// 所有promise都成功，则返回RESOLVED
			case value := <-valueChan:
				newPromise.value = value
				newPromise.currentState = RESOLVED
				return
			}
		}
	}()

	return newPromise
}
