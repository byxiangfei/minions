package minions

import (
	"errors"
	"time"
)

const (
	TaskQueueLen = 10000
	CheckTime    = time.Second
	AddTimeout   = 5 * time.Second
)

type ITask interface {
	Run(interface{}) error
}

var TaskQueue chan ITask

func init() {
	TaskQueue = make(chan ITask, TaskQueueLen) // overall task limiter
}

type Boss struct {
	pool          chan chan ITask
	minionNums    int // max minions nums
	minions       []*Minion
	minionCapcity int // each minion's capacity
	isStop        bool
	quit          chan bool
}

func NewBoss(minionNums int) *Boss {
	return &Boss{
		pool:       make(chan chan ITask, minionNums),
		minionNums: minionNums,
	}
}

func (b *Boss) Start() error {
	for i := 0; i < b.minionNums; i++ {
		m := NewMinion(i, b.pool)
		b.minions = append(b.minions, m)
		m.Work()
	}
	go b.dispatch()
	go b.check()
	return nil
}

func (b *Boss) dispatch() {
	for {
		select {
		case task := <-TaskQueue:
			go func(task ITask) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				// 对minion来说是层保护，每个 minion 里没有队列
				taskCh := <-b.pool
				// dispatch the job to the worker job channel
				taskCh <- task
			}(task)
		case <-b.quit:
			b.Stop()
			return
		}
	}
}

func (b *Boss) check() {
	for {
		<-time.After(CheckTime)
		for _, m := range b.minions {
			if b.isStop {
				b.Stop()
				return
			}
			if m.IsStop() {
				m.Work()
			}
		}
	}
}

func (b *Boss) Stop() error {
	b.isStop = true
	for _, m := range b.minions {
		if !m.IsStop() {
			if err := m.Stop(); err != nil {
				return nil
			}
		}
	}
	return nil
}

func (b *Boss) AddTask(task ITask) error {
	select {
	case TaskQueue <- task:
	case <-time.After(AddTimeout):
		errors.New("Add task time out.")
	}
	return nil
}
