package server

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

type workerJob struct {
	wg *sync.WaitGroup
}

func (w *workerJob) Run() {
	//fmt.Println("Ran")
	//time.Sleep(time.Second)
	w.wg.Done()
}

func BenchmarkWorker_Signal(b *testing.B) {
	pool := NewWorkerPool(context.Background(), 4000, 100000)

	wg := &sync.WaitGroup{}
	job := &workerJob{wg}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		pool.Dispatch(job)
	}

	wg.Wait()
	b.StopTimer()

	b.ReportAllocs()
	pool.PrintStats()

	//pool.Stop()
}

func TestWorker_Add(t *testing.T) {
	pool := NewWorkerPool(context.Background(), 0, 1024)

	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		pool.Dispatch(&workerJob{wg})
		//pool.Get().dispatch(&workerJob{wg})
	}

	wg.Wait()

	pool.PrintStats()

	pool.Stop()
}

func TestNewWorkerPool(t *testing.T) {
	fmt.Printf("%012d | %10s| \n", 12, "hello there today")
}
