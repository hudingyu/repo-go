package actor

import (
	"context"
	"repo-go/promise"
)

type Request struct {
	task   promise.TaskFunc
	result chan *promise.Promise
}

type Actor struct {
	queue  chan *Request
	ctx    context.Context
	cancel context.CancelFunc
}

type TaskFunc func() (interface{}, error)

func NewActor(buffer int) *Actor {
	cancelCtx, cancel := context.WithCancel(context.Background())
	actor := &Actor{
		queue:  make(chan *Request, buffer),
		ctx:    cancelCtx,
		cancel: cancel,
	}

	// 开启执行协程
	go actor.schedule()
	return actor
}

func (actor *Actor) schedule() {
loop:
	for {
		select {
		// 队列中有待处理的请求，则创建新的promise执行请求
		case request := <-actor.queue:
			promise := promise.NewPromise(request.task)
			request.result <- promise
		// actor被cancel，结束
		case <-actor.ctx.Done():
			break loop
		}
	}
}

// 开启一个任务
func (actor *Actor) Do(task promise.TaskFunc) *promise.Promise {
	// 如果actor已经关闭，则返回包含错误的promise
	if actor.ctx.Err() != nil {
		return promise.NewPromise(
			func() (interface{}, error) {
				return nil, actor.ctx.Err()
			})
	}

	// 将任务包装成请求，加入到执行队列，同时立即返回
	request := &Request{task: task, result: make(chan *promise.Promise)}
	actor.queue <- request
	// 返回一个promise 作为future对象，方便处理请求结果
	return promise.NewPromise(
		func() (interface{}, error) {
			p := <-request.result
			return p.Done()
		})
}

// 关闭actor
func (actor *Actor) Close() {
	actor.cancel()
}
