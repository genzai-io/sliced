package server

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/redcon"
	"github.com/genzai-io/sliced/common/service"
	"github.com/rcrowley/go-metrics"
)

type CmdServer struct {
	service.BaseService

	evsrv evio.Server
	wg    sync.WaitGroup

	action evio.Action

	totalConns    metrics.Counter // counter for total connections
	totalCommands metrics.Counter // counter for total commands
	totalBytesIn  metrics.Counter
	totalBytesOut metrics.Counter
}

func NewCmdServer() *CmdServer {
	e := &CmdServer{
		totalConns:    metrics.NewCounter(),
		totalCommands: metrics.NewCounter(),
		totalBytesIn:  metrics.NewCounter(),
		totalBytesOut: metrics.NewCounter(),
	}
	e.BaseService = *service.NewBaseService(moved.Logger, "srv", e)
	return e
}

func (e *CmdServer) OnStart() error {
	e.wg.Add(1)
	go e.serve()
	return nil
}

func (e *CmdServer) OnStop() {
	e.action = evio.Shutdown
	//e.Ev.Shutdown()
}

func (e *CmdServer) Wait() {
	e.wg.Wait()
}

func (e *CmdServer) serve() {
	defer e.wg.Done()

	var events evio.Events

	// Set the number of loops to fire up
	events.NumLoops = moved.EventLoops

	// Try to balance across the event loops
	events.LoadBalance = evio.LeastConnections

	events.Serving = func(srv evio.Server) (action evio.Action) {
		e.evsrv = srv
		return
	}

	events.Opened = func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		// Create new Conn
		co := &Conn{
			evc:     c,
			handler: api.Handler,
		}

		// Let's reuse the read buffer
		opts.ReuseInputBuffer = true

		// Set the context
		c.SetContext(co)

		return
	}

	events.Tick = func() (delay time.Duration, action evio.Action) {
		delay = time.Hour
		if e.action == evio.Shutdown {
			action = evio.Shutdown
		}
		return
	}

	events.Closed = func(co evio.Conn, err error) (action evio.Action) {
		// Notify connection.
		co.Context().(*Conn).closed()
		return
	}

	events.Detached = func(co evio.Conn, rwc io.ReadWriteCloser) (action evio.Action) {
		return
	}

	events.Data = func(co evio.Conn, in []byte) (out []byte, action evio.Action) {
		if co == nil {
			action = evio.Shutdown
			return
		}
		c, ok := co.Context().(*Conn)
		if !ok {
			action = evio.Close
			return
		}

		// Are we woke?
		if in == nil {
			if ctx := c.woke(); ctx != nil {
				return ctx.Out, c.action
			} else {

				//action = evio.Close
				return
			}
		}

		//s.bytesIn.Inc(int64(len(in)))

		// Does the connection have some news to tell the event loop?
		if c.action != evio.None {
			action = c.action
			return
		}

		atomic.AddUint64(&c.statsTotalUpstream, uint64(len(in)))

		// A single buffer is reused at the eventloop level.
		// If we get partial commands then we need to copy to
		// an allocated buffer at the connection level.
		// Zero copy if possible strategy.
		data := c.begin(in)

		//var packet []byte
		var complete bool
		var err error
		//var args [][]byte
		var command api.Command
		var cmdCount = 0

		ctx := &api.Context{
			Conn: c,
			Out:  out,
		}

	LOOP:
		for action == evio.None {
			// Read next command.
			ctx.Packet, complete, ctx.Args, _, data, err = redcon.ParseNextCommand(data, ctx.Args[:0])

			if err != nil {
				action = evio.Close
				out = redcon.AppendError(out, err.Error())
				break
			}

			// Do we need more data?
			if !complete {
				// Exit loop.
				break LOOP
			}

			numArgs := len(ctx.Args)
			if numArgs > 0 {
				c.statsTotalCommands++
				cmdCount++

				//ctx.Name = *(*string)(unsafe.Pointer(&ctx.Args[0]))
				ctx.Name = strings.ToUpper(string(ctx.Args[0]))

				if numArgs > 1 {
					ctx.Key = string(ctx.Args[1])
					//ctx.Extract = *(*string)(unsafe.Pointer(&ctx.Args[1]))
				} else {
					ctx.Key = ""
				}

				before := len(ctx.Out)

				// Parse Command
				command = c.handler.Parse(ctx)

				if command == nil {
					command = cmd.ERR(fmt.Sprintf("command '%s' not found", ctx.Name))
				}

				// Early return?
				if len(ctx.Out) == before {
					c.Invoke(ctx, command)
				}

				if command.IsChange() {
					ctx.AddChange(command, ctx.Packet)
				} else {
					// Commit if necessary
					ctx.Commit()
				}

				ctx.Index++
			}
		}

		// Copy partial Cmd data if any.
		c.end(data)

		atomic.AddUint64(&c.statsTotalDownstream, uint64(len(out)))

		if action == evio.Close {
			return
		}

		if command == nil {
			return ctx.Out, c.action
		}

		// Flush commit buffer
		ctx.Commit()

		return ctx.Out, c.action
	}

	err := evio.Serve(events, fmt.Sprintf("tcp://0.0.0.0:%d?reuseport=true", moved.ApiPort))
	if err != nil {
		e.Logger.Error().Err(err)
	}
}
