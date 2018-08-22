package minions

import (
	"errors"
	"runtime"
	"sync/atomic"
	"time"

	"code.byted.org/gopkg/pkg/log"
)

const (
	StopTimeout    = 10 * time.Second
	ErrorSleepTime = 5 * time.Millisecond
)

type Minion struct {
	id       int
	isStop   int32
	bossPool chan chan ITask // for let boss know you are ready,
	taskChan chan ITask      // to receive boss's new task
	stop     chan struct{}   // receive stop signal
}

func NewMinion(id int, bossPool chan chan ITask) *Minion {
	return &Minion{
		id:       id,
		bossPool: bossPool,
		taskChan: make(chan ITask),
	}
}

func (m *Minion) Work() {
	//if atomic.CompareAndSwapInt32(&m.isStop, 0, 1) {
	go m.run()
	//}
}
func (m *Minion) run() {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 20
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error("worker=%d panic=%v\n%s\n", m.id, err, buf)
		}
		atomic.StoreInt32(&m.isStop, 1)
	}()
	for {
		//将当前Minion中的chan注册到所属的BossPool, 让boss知道可以往里面塞task
		m.bossPool <- m.taskChan
		select {
		case <-m.stop:
			atomic.StoreInt32(&m.isStop, 1)
			return
		case task := <-m.taskChan:
			if err := task.Run(m.id); err != nil {
				time.Sleep(ErrorSleepTime)
			}
		}
	}
}

// IsStop can let Boss whether this minion off the work
func (m *Minion) IsStop() bool {
	return atomic.LoadInt32(&m.isStop) == 1
}

func (m *Minion) Stop() error {
	if m.IsStop() {
		return nil
	}
	select {
	case m.stop <- struct{}{}:
	case <-time.After(StopTimeout):
		return errors.New("stop minion time out")
	}
	return nil
}
