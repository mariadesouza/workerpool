/*
Package workerpool ...

Buffered channels and worker Pools

Worker pools are a model in which a fixed number of m workers have to do n number of work  n tasks in a work pool

Workers: Perform units of work.

Collector: Channel to receive Work requests

availableWorkers Pool:
	Buffered channel of channels used by workers to process work requests

Dispatcher: Pulls work requests off the Collector and sends them to available Channels
*/
package workerpool

const (
	maxWorkers = 15
)

// WorkerPool is a model where a fixed number of m workers have to do n number of work  n tasks in a work pool.
type WorkerPool struct {
	numOfWorkers      int
	availableWorkers  chan chan func()
	collector         chan func()
	quit              chan bool
	activeWorkerCount int
}

// New : Create the workerPool
func New(numOfWorkers int) *WorkerPool {
	// There must be at least one worker.
	if numOfWorkers < 1 {
		numOfWorkers = 1
	}

	//initialize workerPool
	pool := &WorkerPool{
		numOfWorkers:      numOfWorkers,
		collector:         make(chan func()),
		availableWorkers:  make(chan chan func(), maxWorkers),
		quit:              make(chan bool),
		activeWorkerCount: 0,
	}

	// Start the dispatcher.
	go pool.dispatcher()

	return pool
}

// AddWorkToPool : Add the desired concurrent function to the work pool passed as a param
func (p *WorkerPool) AddWorkToPool(task func()) {
	if task != nil {
		p.collector <- task
	}
}

//Close : Gracefully end the workpool once work is all done
func (p *WorkerPool) Close() {
	if p.hasWorkStopped() {
		return
	}
	close(p.collector)
	<-p.quit
}

func (p *WorkerPool) hasWorkStopped() bool {
	select {
	case <-p.quit:
		return true
	default:
	}
	return false
}

// 1.
// dispatcher sends the next task to an idle worker.
func (p *WorkerPool) dispatcher() {
	defer close(p.quit)
	var task func()
	var ok bool
	var workerTaskChan chan func()

DoLoop:
	for {
		select {
		case task, ok = <-p.collector:
			if ok { // we've got work to do
				select {
				case workerTaskChan = <-p.availableWorkers:
					// A worker is ready, so give the task to worker.
					workerTaskChan <- task
				default:
					// No workers available, create a new worker and give task.
					p.createWorker(task)
				}
			} else {
				break DoLoop
			}
		}
	}
	// Stop all remaining workers as they become ready.
	for p.activeWorkerCount > 0 {
		workerTaskChan = <-p.availableWorkers
		close(workerTaskChan)
		p.activeWorkerCount--
	}
}

//.2 create a new worker and give task.
func (p *WorkerPool) createWorker(task func()) {
	availablePool := make(chan chan func())

	if p.activeWorkerCount < p.numOfWorkers {
		p.activeWorkerCount++
		go func(t func()) {
			kickStartWork(availablePool, p.availableWorkers)
			taskChan := <-availablePool
			taskChan <- t
		}(task)
	} else {
		go func(t func()) {
			taskChan := <-p.availableWorkers
			taskChan <- t
		}(task)
	}
}

// 3.
func kickStartWork(availablePool, readyWorkers chan chan func()) {
	go func() {
		taskChan := make(chan func())
		var task func()
		var ok bool
		// Register availability on availablePool.
		availablePool <- taskChan
		for {
			// Read task from dispatcher.
			task, ok = <-taskChan
			if !ok {
				break
			}
			task() // Do the work
			readyWorkers <- taskChan
		}
	}()
}
