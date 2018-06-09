package server

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/service"
	"github.com/rcrowley/go-metrics"
	"github.com/genzai-io/sliced/app/api"
)

type CmdServer struct {
	service.BaseService

	evsrv evio.Server
	wg    sync.WaitGroup

	action evio.Action

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
	e.BaseService = *service.NewBaseService(moved.Logger, "cmd-srv", e)
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

func (e *CmdServer) serve() {
	defer e.wg.Done()

	var events evio.Events

	// Set the number of loops to fire up
	events.NumLoops = moved.EventLoops

	// Try to balance across the event loops
	events.LoadBalance = evio.LeastConnections

	// Fired when it is available to receive connections
	events.Serving = func(srv evio.Server) (action evio.Action) {
		e.evsrv = srv
		return
	}

	// Fired when a new connection is created
	events.Opened = func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		// Create new CmdConn
		// This type of Conn can be upgraded to various other types
		co := &Conn{
			Ev: c,
			Out: &emptyBuffer,
		}

		// Let's reuse the read buffer
		opts.ReuseInputBuffer = true

		// Set the context
		c.SetContext(co)

		e.statsActiveConns.Inc(1)
		e.statsConns.Inc(1)
		e.statsOpened.Inc(1)

		return
	}

	// Periodic event invoked from the Event Loop goroutine / thread
	events.Tick = func() (delay time.Duration, action evio.Action) {
		delay = time.Second * 5
		if e.action == evio.Shutdown {
			action = evio.Shutdown
		}
		return
	}

	// Fired when a connection managed by one of the Event Loops was closed
	events.Closed = func(co evio.Conn, err error) (action evio.Action) {
		e.statsActiveConns.Dec(1)
		e.statsClosed.Inc(1)

		// Notify connection.
		ctx := co.Context()
		if ctx != nil {
			if conn, ok := ctx.(*Conn); ok {
				conn.OnClosed()
				co.SetContext(nil)
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
		c, ok := co.Context().(api.EvData)
		if !ok {
			action = evio.Close
			return
		}

		e.statsIngress.Inc(int64(len(in)))

		return c.OnData(in)
	}

	err := evio.Serve(events, fmt.Sprintf("tcp://0.0.0.0:%d?reuseport=true", moved.ApiPort))
	if err != nil {
		e.Logger.Error().Err(err)
	}
}
