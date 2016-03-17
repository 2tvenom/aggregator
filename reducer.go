package aggregator

import (
	"sync"
	"sync/atomic"
)

type (
	reducer struct {
		reduceChan chan AggregatorTask
		exitChan   chan bool
		task       AggregatorTask
		wait       *sync.WaitGroup
		mutex      *sync.Mutex

		statReduceErrors     uint64
	}
)

func newReducer(task AggregatorTask, reduceQueueMaxLen int) *reducer {
	return &reducer{
		reduceChan: make(chan AggregatorTask, reduceQueueMaxLen),
		exitChan:   make(chan bool, 1),
		task:       task,
		wait:       &sync.WaitGroup{},
		mutex:      &sync.Mutex{},
	}
}

func (r *reducer) GetReduceQueue() chan AggregatorTask {
	return r.reduceChan
}

func (r *reducer) Done() {
	r.exitChan <- true
	r.wait.Wait()
}

func (r *reducer) Start() {
	r.wait.Add(1)
	go r.start()
}

func (r *reducer) start() {
	defer r.wait.Done()
	for {
		select {
		case data := <-r.reduceChan:
			r.mutex.Lock()
			err := r.task.Reduce(data)

			if err != nil {
				atomic.AddUint64(&r.statReduceErrors, 1)
			}

			r.mutex.Unlock()
		case <-r.exitChan:
			for len(r.reduceChan) > 0 {
				r.mutex.Lock()
				err := r.task.Reduce(<-r.reduceChan)

				if err != nil {
					atomic.AddUint64(&r.statReduceErrors, 1)
				}

				r.mutex.Unlock()
			}
			return
		}
	}
}