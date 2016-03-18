package aggregator

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	workerPool struct {
		dataSource chan interface{}
		reduceChan chan AggregatorTask
		exitChan   chan bool
		task       AggregatorTask
		wait       *sync.WaitGroup
		mutex      *sync.Mutex

		maxWorkers       int
		reduceMaxCounter int

		statPreProcessErrors uint64
		statMapErrors        uint64
		statProcessed        uint64
	}
)

func newWorkerPool(task AggregatorTask, dataSource chan interface{}, reduceChan chan AggregatorTask, maxWorkers int, entityForReduce int) *workerPool {
	return &workerPool{
		dataSource:       dataSource,
		reduceChan:       reduceChan,
		maxWorkers:       maxWorkers,
		reduceMaxCounter: entityForReduce,
		exitChan:         make(chan bool, maxWorkers),
		task:             task,
		wait:             &sync.WaitGroup{},
		mutex:            &sync.Mutex{},
	}
}

func (p *workerPool) Start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wait.Add(1)
		go p.worker(i)
	}
}

func (p *workerPool) Done() {
	for i := 0; i < p.maxWorkers; i++ {
		p.exitChan <- true
	}
	p.wait.Wait()
}

func (p *workerPool) worker(id int) {
	defer p.wait.Done()

	w := newWorker(p.reduceChan, p.task, p.reduceMaxCounter)
	for {
		select {
		case data := <-p.dataSource:
			w.process(data)
		case <-time.Tick(time.Second * 10):
			for len(p.dataSource) > 0 {
				w.process(<-p.dataSource)
			}

			w.flushCache()
		case <-p.exitChan:
			for len(p.dataSource) > 0 {
				w.process(<-p.dataSource)
			}

			w.flushCache()

			atomic.AddUint64(&p.statPreProcessErrors, w.statPreProcessErrors)
			atomic.AddUint64(&p.statMapErrors, w.statMapErrors)
			atomic.AddUint64(&p.statProcessed, w.statProcessed)

			return
		}
	}
}