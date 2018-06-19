/*Package workerpool ...

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

//Close : Gracefully end the workpool once work is all done
func (p *WorkerPool) Close() {
	if p.hasWorkStopped() {
		return
	}
	for p.activeWorkerCount > 0 {
		workerTaskChan := <-p.availableWorkers
		close(workerTaskChan)
		p.activeWorkerCount--
	}
	close(p.collector)
	<-p.quit
}

// AddWorkToPool : Add the desired concurrent function to the work pool passed as a param
func (p *WorkerPool) AddWorkToPool(task func()) {
	if task != nil {
		p.collector <- task
	}
}

// Stopped returns true if this worker pool has been stopped.
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

	var task func()
	var ok bool
	var workerTaskChan chan func()

DoLoop:
	for {
		p := p
		select {
		case task, ok = <-p.collector:
			if ok { // we got work to do
				select {
				case workerTaskChan = <-p.availableWorkers:
					// A worker is ready, so give task to worker.
					workerTaskChan <- task
				default:
					// NO ready workers, create a new worker from pool.
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
	close(p.quit)
}

//.2
func (p *WorkerPool) createWorker(task func()) {
	availablePool := make(chan chan func())

	if p.activeWorkerCount < p.numOfWorkers {
		p.activeWorkerCount++
		go func(t func()) {
			startWork(availablePool, p.availableWorkers)
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
func startWork(availablePool, readyWorkers chan chan func()) {
	go func() {
		taskChan := make(chan func())
		var task func()
		var ok bool
		// Register availability on availablePool .
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
