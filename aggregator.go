package aggregator

import (
	"fmt"
	"runtime"
	"time"
)

type (
	AggregatorTask interface {
		Source(chan interface{})
		GetBlank() AggregatorTask
		Map(interface{}) error
		Reduce(AggregatorTask) error
	}

	AggregatorSource interface {
		Source(chan interface{})
	}

	AggregatorEntityPreProcess interface {
		EntityPreProcess(interface{}) (interface{}, error)
	}

	AggregatorTaskPostProcess interface {
		PostProcess() error
	}

	Aggregator struct {
		tasks              []AggregatorTask
		maxGoRoutines      int
		maxQueueLen        int
		maxReduceQueueLen  int
		maxEntityForReduce int

		statReduceErrors     uint64
		statPreProcessErrors uint64
		statMapErrors        uint64
		statProcessed        uint64
		executionTime        time.Duration
	}
)

const (
	DefaultMaxGoRoutines      = 1
	DefaultMaxQueue           = 100
	DefaultMaxReduceQueue     = 10
	DefaultMaxEntityForReduce = 100
)

func NewAggregator() *Aggregator {
	return &Aggregator{
		tasks:              []AggregatorTask{},
		maxGoRoutines:      DefaultMaxGoRoutines,
		maxQueueLen:        DefaultMaxQueue,
		maxReduceQueueLen:  DefaultMaxReduceQueue,
		maxEntityForReduce: DefaultMaxEntityForReduce,
	}
}

func (a *Aggregator) SetMaxGoRoutines(quantity int) {
	a.maxGoRoutines = quantity
}

func (a *Aggregator) MaxGoRoutines() int {
	return a.maxGoRoutines
}

func (a *Aggregator) SetMaxQueueLen(quantity int) {
	a.maxQueueLen = quantity
}

func (a *Aggregator) MaxQueueLen() int {
	return a.maxQueueLen
}

func (a *Aggregator) SetMaxReduceQueueLen(quantity int) {
	a.maxReduceQueueLen = quantity
}

func (a *Aggregator) MaxReduceQueueLen() int {
	return a.maxReduceQueueLen
}

func (a *Aggregator) SetMaxEntityForReduce(quantity int) {
	a.maxEntityForReduce = quantity
}

func (a *Aggregator) MaxEntityForReduce() int {
	return a.maxEntityForReduce
}

func (a *Aggregator) CountReduceErrors() uint64 {
	return a.statReduceErrors
}

func (a *Aggregator) CountPreProcessErrors() uint64 {
	return a.statPreProcessErrors
}

func (a *Aggregator) CountMapErrors() uint64 {
	return a.statMapErrors
}

func (a *Aggregator) CountProcessed() uint64 {
	return a.statProcessed
}

func (a *Aggregator) ExecutionTime() time.Duration {
	return a.executionTime
}

func (a *Aggregator) AddTask(t AggregatorTask) {
	a.tasks = append(a.tasks, t)
}

func (a *Aggregator) Start() {
	//requiredCpu := a.maxGoRoutines + 2
	//if runtime.NumCPU() < requiredCpu {
	//	panic(fmt.Sprintf("Requered %d CPU (MaxGoRoutines + 2)", requiredCpu))
	//}

	start := time.Now()

	for _, task := range a.tasks {
		sourceData := make(chan interface{}, a.maxQueueLen)
		reducer := newReducer(task, a.maxReduceQueueLen)
		workerPool := newWorkerPool(task, sourceData, reducer.GetReduceQueue(), a.maxGoRoutines, a.maxEntityForReduce)

		workerPool.Start()
		reducer.Start()

		task.Source(sourceData)

		workerPool.Done()
		reducer.Done()

		if taskPostProcess, ok := task.(AggregatorTaskPostProcess); ok {
			taskPostProcess.PostProcess()
		}

		a.statReduceErrors += reducer.statReduceErrors
		a.statPreProcessErrors += workerPool.statPreProcessErrors
		a.statMapErrors += workerPool.statMapErrors
		a.statProcessed  += workerPool.statProcessed
	}
	a.executionTime = time.Now().Sub(start);
}
