package core

import (
	"context"
	"sync"
	"time"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/proto/store"
	"github.com/slice-d/genzai/common/service"
)

type Roller struct {
	service.BaseService

	cycle  uint64
	model  *store.Roller
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	topics map[int64]*Topic
}

func newRoller(model *store.Roller) *Roller {
	ctx, cancel := context.WithCancel(context.Background())
	roller := &Roller{
		ctx:    ctx,
		cancel: cancel,
		topics: make(map[int64]*Topic),
	}

	roller.BaseService = *service.NewBaseService(moved.Logger, "roller."+model.Name, roller)

	return roller
}

func (r *Roller) OnStart() error {
	// Run the timer goroutine
	r.wg.Add(1)
	go r.runTimer()

	return nil
}

func (r *Roller) OnStop() {
	r.cancel()
	r.wg.Wait()
}

func (r *Roller) OnReset() error {
	return nil
}

func (r *Roller) runTimer() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return

		case <-time.After(time.Second):

		}
	}
}

// Returns the current cycle number
func (r *Roller) CurrentCycle() uint64 {
	return r.cycle
}
