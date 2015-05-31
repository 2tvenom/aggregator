# Golang helper for parallel data processing
----------
Simple helper for fast starting processing big files/tables/logs/etc. This code has been developed and maintained by Ven at May 2015.

## How it use

You need creatr struct for this interface:
``` go 
	AggregatorTask interface {
		Source(chan interface{})
		GetBlank() AggregatorTask
		Map(interface{}) error
		Reduce(AggregatorTask) error
	}
```

`Source(chan interface{})`
This is source of your data. Just put read it and put to chan
`GetBlank() AggregatorTask` 
Rerurn new prepared struct. Goroutine will use it for local cache and after put it to reduce into current struct. _Not pointer to current struct._
`Map(interface{}) error`
Process your data
`Reduce(AggregatorTask) error`
Merge local cache from goroutine to main struct

## Example 

```go
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

//Source or your data. There you can downlaod file by http, read from disk or table. As you want. Just put to chan
func (e *entity) Source(source chan interface{}) {
	var i uint64
	for i=0; i<100; i++ {
		source <- i
	}
}

//Return blank struct (For local cache of goroutine)
func (e *entity) GetBlank() aggregator.AggregatorTask {
	return &entity{}
}

//Process data
func (e *entity) Map(data interface{}) error {
	e.count += data.(uint64)
	return nil
}

//Merge data from local goroutine cache to main struct
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

```

## Preferences
`SetMaxGoRoutines(quantity int)`
Count of goroutines for parallel data processing _(Default = 1)_

`SetMaxQueueLen(quantity int)`
Max length of source queue chan _(Default = 100)_

`SetMaxReduceQueueLen(quantity int)`
Max length of reduce queue  _(Default = 10)_

`SetMaxEntityForReduce(quantity int)`
How many process entity before send local cache to reduce   _(Default = 100)_

## Statistic
Method for receiving erorrs

`CountReduceErrors()`
`CountPreProcessErrors()`
`CountMapErrors()`
`CountProcessed()`
