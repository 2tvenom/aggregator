package main

import (
	"github.com/2tvenom/aggregator"
	"fmt"
	"time"
	"math/rand"
)

type (
	entity struct {
		count uint64
	}
)

func (e *entity) Source(source chan interface{}) {
	var i uint64
	for i=0; i<100; i++ {
		source <- i
	}
}

func (e *entity) GetBlank() aggregator.AggregatorTask {
	return &entity{}
}


func (e *entity) Map(data interface{}) error {
	e.count += 1
	<-time.After(time.Millisecond * time.Duration(rand.Intn(3) * 50))
	return nil
}

func (e *entity) Reduce(localCache aggregator.AggregatorTask) error {
	e.count += localCache.(*entity).count
	return nil
}

func main() {
	entityTask := &entity{}

	a := aggregator.NewAggregator()
	a.SetMaxEntityForReduce(3)
	a.AddTask(entityTask)
	a.Start()


	fmt.Printf("%d\n", entityTask.count)
}
