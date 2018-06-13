package server

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/service"
	"github.com/genzai-io/sliced/common/spinlock"
	"github.com/pointc-io/sliced/index/celltree"
	"github.com/rcrowley/go-metrics"
)

type CmdServer struct {
	service.BaseService

	evsrv evio.Server
	wg    sync.WaitGroup

	action evio.Action

	loops []*eventLoop

	//connections    *btree.BTree
	//connectionLock spinlock.Locker

	connectionCount uint64

	statsConns       metrics.Counter // counter for total connections
	statsOpened      metrics.Counter
	statsClosed      metrics.Counter
	statsCommands    metrics.Counter // counter for total commands
	statsIngress     metrics.Counter
	statsEgress      metrics.Counter
	statsActiveConns metrics.Counter
	statsWakes       metrics.Counter
}

func NewCmdServer() *CmdServer {
	e := &CmdServer{
		statsConns:       metrics.NewCounter(),
		statsOpened:      metrics.NewCounter(),
		statsClosed:      metrics.NewCounter(),
		statsCommands:    metrics.NewCounter(),
		statsIngress:     metrics.NewCounter(),
		statsEgress:      metrics.NewCounter(),
		statsActiveConns: metrics.NewCounter(),
		statsWakes:       metrics.NewCounter(),
	}
	//e.connections = btree.New(64, e)
	e.BaseService = *service.NewBaseService(moved.Logger, "svr", e)
	return e
}

func (e *CmdServer) OnStart() error {
	e.wg.Add(1)
	go e.serve()
	return nil
}

func (e *CmdServer) OnStop() {
	e.action = evio.Shutdown
	if e.evsrv.Shutdown != nil {
		e.evsrv.Shutdown()
		if e.evsrv.WaitForShutdown != nil {
			e.evsrv.WaitForShutdown()
		}
	}
}

func (e *CmdServer) Wait() {
	e.wg.Wait()
}

func (e *CmdServer) Dial(addr string) error {
	return nil
}

func (e *CmdServer) loop(index int) *eventLoop {
	if index < 0 {
		return nil
	} else if index >= len(e.loops) {
		return nil
	} else {
		return e.loops[index]
	}
}

func (e *CmdServer) serve() {
	defer e.wg.Done()

	var events evio.Events

	// Set the number of loops to fire up
	events.NumLoops = moved.EventLoops

	// Create event loops
	e.loops = make([]*eventLoop, events.NumLoops)
	for i := 0; i < events.NumLoops; i++ {
		e.loops[i] = newEventLoop(e)
	}

	// Try to balance evenly across the event loops
	events.LoadBalance = evio.LeastConnections

	// Fired when it is available to receive connections
	events.Serving = func(srv evio.Server) (action evio.Action) {
		e.evsrv = srv
		return
	}

	// Fired when a new connection is created
	events.Opened = func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		nextID := atomic.AddUint64(&e.connectionCount, 1)

		// Create new CmdConn
		// This type of Conn can be upgraded to various other types
		co := &CmdConn{
			ID: nextID,
			ev: c,
			//Out: &emptyBuffer,
		}
		co.onData = co.OnData

		// Let's reuse the read buffer
		opts.ReuseInputBuffer = true

		// Set the context
		c.SetContext(co)

		e.statsActiveConns.Inc(1)
		e.statsConns.Inc(1)
		e.statsOpened.Inc(1)

		// Add to btree
		loop := e.loop(c.LoopIndex())
		if loop != nil {
			loop.connections.Insert(nextID, unsafe.Pointer(co), 0)
		}

		return
	}

	// Periodic event invoked from the Event Loop goroutine / thread
	events.Tick = func(loopIndex int) (delay time.Duration, action evio.Action) {
		if e.action == evio.Shutdown {
			action = evio.Shutdown
			return
		}
		loop := e.loop(loopIndex)
		if loop != nil {
			return loop.tick()
		}
		return
	}

	// Fired when a connection managed by one of the Event Loops was closed
	events.Closed = func(co evio.Conn, err error) (action evio.Action) {
		e.statsActiveConns.Dec(1)
		e.statsClosed.Inc(1)

		// Remove from btree

		// Notify connection.
		ctx := co.Context()
		if ctx != nil {
			if conn, ok := ctx.(*CmdConn); ok {
				conn.OnClosed()
				co.SetContext(nil)

				// Remove from the loop's management
				loop := e.loop(co.LoopIndex())
				if loop != nil {
					loop.remove(conn)
				}
			}
		}
		return
	}

	events.Detached = func(co evio.Conn, rwc io.ReadWriteCloser) (action evio.Action) {
		c, ok := co.Context().(api.EvDetacher)
		if ok {
			c.OnDetach(rwc)
		}
		return
	}

	// Fired when there is new data to be read and/or when notified about data to write.
	// evio.Conn.Wake() will call this method with a nil "in" param signaling it
	// is available to write.
	events.Data = func(co evio.Conn, in []byte) (out []byte, action evio.Action) {
		if co == nil {
			action = evio.Shutdown
			return
		}
		c, ok := co.Context().(*CmdConn)
		if !ok {
			action = evio.Close
			return
		}

		e.statsIngress.Inc(int64(len(in)))

		return c.onData(in)
	}

	err := evio.Serve(events, fmt.Sprintf("tcp://0.0.0.0:%d?reuseport=true", moved.ApiPort))
	if err != nil {
		e.Logger.Error().Err(err)
	}
}

// Only scan this many connections per tick before picking up where
// it left off on the next tick. This ensures the event-loop stays
// very responsive regardless of the number of connections under
// management.
const maxIterationPerTick = 1000
const minTickDuration = time.Millisecond * 10
const maxTickDuration = time.Second

type eventLoop struct {
	spinlock.Locker
	svr         *CmdServer
	connections celltree.Tree

	// Connections that have a live worker goroutine.
	workers celltree.Tree

	pivot uint64
}

func newEventLoop(svr *CmdServer) *eventLoop {
	loop := &eventLoop{
		svr: svr,
	}
	return loop
}

func (l *eventLoop) remove(conn *CmdConn) {
	l.connections.Remove(conn.ID, unsafe.Pointer(conn))
}

func (l *eventLoop) tick() (delay time.Duration, action evio.Action) {
	delay = time.Second

	var (
		count = 0
		conn  *CmdConn
	)

	l.connections.Range(l.pivot, func(cell uint64, key unsafe.Pointer, extra uint64) bool {
		count++
		l.pivot = cell

		if key == nil {
			return true
		}

		conn = (*CmdConn)(key)
		conn.tick()

		return true
	})

	if count == maxIterationPerTick {
		delay = minTickDuration
	} else {
		// Reset
		delay = maxTickDuration
		l.pivot = 0
	}

	return
}
