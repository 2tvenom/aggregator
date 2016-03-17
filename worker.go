package aggregator

type (
	worker struct {
		counter int
		task AggregatorTask
		currentLocalCache AggregatorTask
		preProcess AggregatorEntityPreProcess
		preProcessRequired bool

		maxReduceOneCache int
		reduceChan chan AggregatorTask

		statPreProcessErrors uint64
		statMapErrors uint64
		statProcessed uint64
	}
)

func newWorker(reduceChan chan AggregatorTask, task AggregatorTask, maxReduceOneCache int) *worker {
	w := &worker{
		task: task,
		currentLocalCache: task.GetBlank(),
		maxReduceOneCache: maxReduceOneCache,
		reduceChan: reduceChan,
	}

	w.preProcess, w.preProcessRequired = w.currentLocalCache.(AggregatorEntityPreProcess)

	return w
}

func (p *worker) process(data interface{}) {
	if p.preProcessRequired {
		var err error
		data, err = p.preProcess.EntityPreProcess(data)
		if err != nil {
			p.statPreProcessErrors++
		}
	}

	err := p.currentLocalCache.Map(data)
	if err != nil {
		p.statMapErrors++
		return
	}
	p.statProcessed++
	p.counter++

	if p.counter >= p.maxReduceOneCache {
		p.flushCache()
	}
}

func (p *worker) flushCache() {
	p.reduceChan <- p.currentLocalCache
	p.currentLocalCache = p.task.GetBlank()
	p.counter = 0
}


