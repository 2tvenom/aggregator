package aggregator

import (
	"sync"
	"sync/atomic"
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

	localCache := p.task.GetBlank()
	counter := 0

	preProcess, preProcessRequired := p.task.(AggregatorEntityPreProcess)
	for {
		select {
		case data := <-p.dataSource:
			if preProcessRequired {
				var err error
				data, err = preProcess.EntityPreProcess(data)
				if err != nil {
					atomic.AddUint64(&p.statPreProcessErrors, 1)
				}
			}

			err := localCache.Map(data)
			if err != nil {
				atomic.AddUint64(&p.statMapErrors, 1)
				continue
			}
			atomic.AddUint64(&p.statProcessed, 1)
			counter++

			if counter >= p.reduceMaxCounter {
				p.reduceChan <- localCache
				localCache = p.task.GetBlank()
				counter = 0
			}

		default:
			select {
			case <-p.exitChan:
				p.reduceChan <- localCache
				return
			default:

			}
		}
	}
}
