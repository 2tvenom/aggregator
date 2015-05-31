package main

import (
	"github.com/2tvenom/aggregator"
	"fmt"
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
	e.count += data.(uint64)
	return nil
}

func (e *entity) Reduce(localCache aggregator.AggregatorTask) error {
	e.count += localCache.(*entity).count
	return nil
}

func main() {
	entityTask := &entity{}

	a := aggregator.NewAggregator()
	a.AddTask(entityTask)
	a.Start()


	fmt.Printf("%d\n", entityTask.count)
}
