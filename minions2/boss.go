package minions2

import (
	"errors"
	"time"
)

const (
	TaskQueueLen = 10000
	CheckTime    = time.Second
	AddTimeout   = time.Second
)

type ITask interface {
	Run(interface{}) error
}

type Boss struct {
	minionNums int // max minions nums
	minions    []*Minion
	isStop     bool
	quit       chan bool
	taskQueue  chan ITask
}

func NewBoss(minionNums int) *Boss {
	return &Boss{
		minionNums: minionNums,
		taskQueue:  make(chan ITask, TaskQueueLen),
	}
}

func (b *Boss) Start() error {
	for i := 0; i < b.minionNums; i++ {
		m := NewMinion(i, b.taskQueue)
		b.minions = append(b.minions, m)
		m.Work()
	}
	go b.check()
	return nil
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
	case b.taskQueue <- task:
	case <-time.After(AddTimeout):
		errors.New("Add task time out.")
	}
	return nil
}
