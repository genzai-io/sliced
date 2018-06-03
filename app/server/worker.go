package server

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	Workers = NewWorkerPool(context.Background(), 0, 0)
)

const (
	Idle    int32 = iota
	Active        = 1
	Closing       = 2
	Closed        = 3
)

type WorkerJob interface {
	Run()
}

type _noopJob struct {
	WorkerJob
}

func (j *_noopJob) Run(ctx context.Context) {}

var noopJob = &_noopJob{}
var closeJob = &_noopJob{}

type WorkerPool struct {
	name   string
	ctx    context.Context
	cancel context.CancelFunc
	pool   *sync.Pool
	mu     sync.Mutex
	state  int32
	min    int64
	max    int64

	workers []*Worker

	// Track Worker supply
	supply    metrics.Counter
	maxSupply metrics.Counter
	forged    metrics.Counter
	exhausted metrics.Counter
	gets      metrics.Counter
	puts      metrics.Counter

	wg sync.WaitGroup
}

func (w *WorkerPool) Name() string {
	return w.name
}

func NewWorkerPool(ctx context.Context, min, max int) *WorkerPool {
	if min < 0 {
		min = 0
	}
	if max < min {
		max = min
	}

	c, cancel := context.WithCancel(ctx)

	w := &WorkerPool{
		ctx:     c,
		cancel:  cancel,
		state:   Active,
		mu:      sync.Mutex{},
		min:     int64(min),
		max:     int64(max),
		wg:      sync.WaitGroup{},
		pool:    new(sync.Pool),
		workers: make([]*Worker, min),

		// Metrics
		supply:    metrics.NewCounter(),
		maxSupply: metrics.NewCounter(),
		forged:    metrics.NewCounter(),
		exhausted: metrics.NewCounter(),
		gets:      metrics.NewCounter(),
		puts:      metrics.NewCounter(),
	}

	// Spin up to satisfy min.
	for i := 0; i < min; i++ {
		w.Put(w.Get())
	}

	go w.runEvictor()

	return w
}

func (w *WorkerPool) runEvictor() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(time.Second * 5):
		}
	}
}

func (w *WorkerPool) PrintStats() {
	fmt.Printf("Supply: %d\n", w.supply.Count())
	fmt.Printf("Max Supply: %d\n", w.maxSupply.Count())
	fmt.Printf("Forged: %d\n", w.forged.Count())
	fmt.Printf("Exhausted: %d\n", w.exhausted.Count())
	fmt.Printf("Gets: %d\n", w.gets.Count())
	fmt.Printf("Puts: %d\n", w.puts.Count())
}

func (w *WorkerPool) Register(registry metrics.Registry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	name := "WorkerPool." + w.name
	registry.Register(name+".supply", w.supply)
	registry.Register(name+".maxSupply", w.maxSupply)
	registry.Register(name+".forged", w.forged)
	registry.Register(name+".exhausted", w.exhausted)
	registry.Register(name+".gets", w.gets)
	registry.Register(name+".puts", w.puts)
}

func (w *WorkerPool) Stop() {
	w.mu.Lock()
	if w.state <= Closing {
		w.mu.Unlock()
		w.wg.Wait()
		return
	}

	w.state = Closing
	w.cancel()
	for _, worker := range w.workers {
		worker.stop()
	}
	w.mu.Unlock()

	// Wait for all to stop.
	w.wg.Wait()

	// Flip state to Closed
	w.mu.Lock()
	w.state = Closed
	w.mu.Unlock()
}

func (w *WorkerPool) Get() *Worker {
	w.gets.Inc(1)

	worker := w.pool.Get()
	if worker != nil {
		return worker.(*Worker)
	}

	// Barrier
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state >= Closing {
		return nil
	}

	// Increase the supply.
	w.supply.Inc(1)

	if w.supply.Count() > w.maxSupply.Count() {
		w.maxSupply.Inc(w.supply.Count() - w.maxSupply.Count())
	}

	// Did we exceed the max?
	if w.max > 0 && w.supply.Count() >= w.max {
		// Exhausted.
		w.exhausted.Inc(1)
		// Decrease supply.
		w.supply.Dec(1)
		// Return nil. Out of Memory
		return nil
	}

	// Increase forged count
	w.forged.Inc(1)

	// Create new Worker.
	wkr := &Worker{
		state: Idle,
		pool:  w,
		//cond:  &sync.Cond{L: &sync.Mutex{}},
		//job:  noopJob,
		wg:   sync.WaitGroup{},
		jobs: make(chan WorkerJob, 1),
	}
	wkr.wg.Add(1)

	// Add to tracking slice
	w.workers = append(w.workers, wkr)

	// Track goroutine exit
	w.wg.Add(1)
	// Start worker
	go wkr.run(&w.wg)

	// Return Worker
	return wkr
}

func (w *WorkerPool) Put(worker *Worker) {
	if worker.state != Idle {
		return
	}
	w.puts.Inc(1)
	w.pool.Put(worker)
}

func (w *WorkerPool) Dispatch(job WorkerJob) bool {
	// Allocate a worker
	worker := w.Get()

	// Was there a problem?
	if worker == nil {
		// Is the pool closing?
		if w.state >= Closing {
			return false
		}

		// Let's block momentarily
		time.Sleep(time.Millisecond)
		worker = w.Get()
		if worker == nil {
			return false
		}
	}

	// Dispatch the job
	return worker.dispatch(job)
}

// Maintains a cached goroutine that can process one job at a time
// and is returned back to a sync.Pool
type Worker struct {
	state int32
	pool  *WorkerPool
	//cond  *sync.Cond
	wg sync.WaitGroup
	//job  WorkerJob
	jobs chan WorkerJob
}

func (i *Worker) casState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i.state, old, new)
}

func (w *Worker) stop() {
	if !w.casState(Idle, Closing) && !w.casState(Active, Closing) {
		return
	}
	close(w.jobs)
	w.wg.Wait()
}

func (w *Worker) dispatch(job WorkerJob) bool {
	if !w.casState(Idle, Active) {
		return false
	}
	w.jobs <- job
	return true
}

func (w *Worker) run(wg *sync.WaitGroup) {
	defer wg.Done()
	defer w.wg.Done()

	for w.state < Closing {
		select {
		case j, ok := <-w.jobs:
			// Was the channel closed?
			if !ok {
				atomic.SwapInt32(&w.state, Closed)
				return
			}

			// Run job under pool Context
			j.Run()

			// Switch back to Idle
			atomic.SwapInt32(&w.state, Idle)
			// Put back into sync.Pool
			w.pool.Put(w)
		}
	}

	atomic.SwapInt32(&w.state, Closed)
}

//func (w *Worker) stop() {
//	w.cond.L.Lock()
//	if w.state >= Closing {
//		w.cond.L.Unlock()
//		return
//	}
//
//	w.state = Closing
//	w.job = closeJob
//	//w.job = noopJob
//	w.cond.Broadcast()
//	w.cond.L.Unlock()
//
//	w.wg.Wait()
//
//	// Decrease supply
//	w.pool.supply.Dec(1)
//}
//
//func (w *Worker) SignalM(job WorkerJob) bool {
//	w.cond.L.Lock()
//	if w.state != Idle {
//		w.cond.L.Unlock()
//		return false
//	}
//	w.state = Active
//	w.job = job
//	w.cond.dispatch()
//	w.cond.L.Unlock()
//	return true
//}
//
//func (w *Worker) run(wg *sync.WaitGroup) {
//	defer wg.Done()
//	defer w.wg.Done()
//
//	var job WorkerJob = noopJob
//	for w.state < Closing {
//		w.cond.L.Lock()
//		for w.job == noopJob {
//			w.cond.Wait()
//		}
//		job = w.job
//		w.cond.L.Unlock()
//
//		if w.state >= Closing {
//			w.state = Closed
//			return
//		}
//
//		// Run job.
//		job.Run()
//
//		// Clear
//		//w.cond.L.Lock()
//		w.job = noopJob
//		w.state = Idle
//		//w.cond.L.Unlock()
//
//		// Return to pool.
//		//w.pool.Put(w)
//	}
//
//	w.state = Closed
//}
