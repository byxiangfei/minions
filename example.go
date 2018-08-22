package main

import (
	"fmt"

	"time"

	"github.com/byxiangfei/minions"
	"github.com/byxiangfei/minions2"
)

type DemoJob struct{}

func (d *DemoJob) Run(input interface{}) error {
	fmt.Println("hello world")
	return nil
}

func test() {
	t0 := time.Now()
	defer func(t time.Time) {
		fmt.Println(time.Now().Sub(t))
	}(t0)

	minionNums := 10
	boss := minions.NewBoss(minionNums)
	boss.Start()

	job := &DemoJob{}
	for i := 0; i < 600000; i++ {
		boss.AddTask(job)
	}

	boss.Stop() // 说stop 就 stop 了，
}

func test2() {
	t0 := time.Now()
	defer func(t time.Time) {
		fmt.Println(time.Now().Sub(t))
	}(t0)

	minionNums := 10
	boss := minions2.NewBoss(minionNums)
	boss.Start()

	job := &DemoJob{}
	for i := 0; i < 600000; i++ {
		boss.AddTask(job)
	}

	boss.Stop() // 这个会把所有的都处理完
}

func main() {
	test()
	//test2()
}
