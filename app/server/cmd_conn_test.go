package server

import (
	"github.com/genzai-io/sliced/common/evio"
	"net"
	"testing"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/app/api"
	"sync"
	"context"
	"time"
	"fmt"
	"strings"
	"io"
)

type WorkerSleep = cmd.Sleep

func TestConn_OnData(t *testing.T) {
	conn := newMockConn()

	conn.send(&cmd.Get{Key: "hi"}, &cmd.Get{ Key: "bye"}).wait()
	conn.send(&cmd.Get{Key: "hi"}).wait()
	conn.send(&WorkerSleep{})
	for i := 0; i < 100; i++ {
		conn.send(&cmd.Get{Key: "hi"})
	}
	conn.send(&cmd.Get{Key: "hi"}).wait()
	// Parse responses
	//conn.send(&cmd.Get{Key: "hello"}).wait()
	//conn.send(&cmd.Set{Key: "hi", Value: "bye"}).wait()

	conn.close()

	//fmt.Print(string(out))
}

func newMockConn() *mockEvConn {
	ctx, cancel := context.WithCancel(context.Background())
	ev := &mockEvConn{
		ctx:         ctx,
		cancel:      cancel,
		eventLoopCh: make(chan interface{}, 2),
		reader:      api.NewReplyReader([]byte{}),
		//dataCh:      make(<-chan *loopEvent),
		//sendCh:      make(chan *requestGroup, 1),
		//recvCh:      make(chan *requestReply, 1),
	}
	conn := newConn(ev)
	ev.conn = conn
	ev.onData = conn.OnData

	ev.start()
	return ev
}

type mockEvConn struct {
	evio.Conn

	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex
	wg     sync.WaitGroup
	conn   *Conn
	onData func(in []byte) (out []byte, action evio.Action)

	closed     bool
	seq        int64
	queue      []*requestGroup
	dataEvents []*loopEvent

	replyIndex int
	requests   []*requestReply

	eventLoopCh chan interface{}

	//dataCh <-chan *loopEvent
	//sendCh chan *requestGroup
	//recvCh chan *requestReply

	reader *api.ReplyReader
}

type expectation int

const (
	onEventLoop expectation = iota
	onWorker
)

type loopCond struct {

}

type loopEvent struct {
	sequence int64
	in       []byte
	out      []byte
	action   evio.Action

	requests []*requestReply
	replies []api.CommandReply
}

func (l *loopEvent) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\nEvent: %d\n", l.sequence + 1))

	if len(l.in) == 0 {
		if len(l.out) == 0 {
			builder.WriteString("===============================\n")
			builder.WriteString("WAKE\n")
			builder.WriteString("===============================\n\n")
		} else {
			builder.WriteString("===============================\n")
			builder.WriteString("WAKE\n")
			builder.WriteString("-------------------------------\n")
			builder.WriteString(fmt.Sprintf("%d RESPONSE(s) [%d bytes]\n", len(l.replies), len(l.out)))
			builder.WriteString("-------------------------------\n")
			builder.WriteString(string(l.out))

			if len(l.requests) > len(l.replies) {
				builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests) - len(l.replies)))
			}

			builder.WriteString("===============================\n\n")
		}
	} else {
		builder.WriteString("===============================\n")
		builder.WriteString(fmt.Sprintf("%d REQUEST(s) [%d bytes]\n", len(l.requests), len(l.in)))
		builder.WriteString("-------------------------------\n")
		builder.WriteString(string(l.in))
		builder.WriteString("-------------------------------\n")

		if len(l.out) == 0 {
			builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests) - len(l.replies)))
			builder.WriteString("===============================\n\n")
		} else {
			builder.WriteString(fmt.Sprintf("%d RESPONSE(s) [%d bytes]\n", len(l.replies), len(l.out)))
			builder.WriteString("-------------------------------\n")
			builder.WriteString(string(l.out))

			if len(l.requests) > len(l.replies) {
				builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests) - len(l.replies)))
			}
			builder.WriteString("===============================\n\n")
		}
	}

	return builder.String()
}

type requestGroup struct {
	requests []*requestReply
	in       []byte
	out      []byte

	replyCounter int

	chResponse chan interface{}
}

func (r *requestGroup) wait() *requestGroup {
	if len(r.requests) == 0 {
		return r
	}
	for _, request := range r.requests {
		request.response()
		r.replyCounter++
	}
	close(r.chResponse)
	return r
}

type requestReply struct {
	command api.Command
	request []byte
	reply   api.CommandReply

	expectation expectation

	chResponse chan api.CommandReply
}

func (r *requestReply) response() *requestReply {
	select {
	case reply, ok := <-r.chResponse:
		if !ok {
			return r
		}
		r.reply = reply
		close(r.chResponse)
		return r

	case <-time.After(10 * time.Second):
		close(r.chResponse)
		return r
	}
	return r
}

func (c *mockEvConn) send(commands ...api.Command) *requestGroup {
	group := &requestGroup{
		chResponse: make(chan interface{}, 1),
	}
	for _, command := range commands {
		request := &requestReply{
			command:    command,
			request:    command.Marshal(nil),
			chResponse: make(chan api.CommandReply, 1),
		}
		group.requests = append(group.requests, request)
		group.in = append(group.in, request.request...)
	}

	c.eventLoopCh <- group
	return group
}

func (c *mockEvConn) expect() {
	//redcon.Parse()
}

type closeMsg struct{}

func (c *mockEvConn) start() {
	// event loop
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {

			case <-c.ctx.Done():
				return

			case event, ok := <-c.eventLoopCh:
				if !ok {
					return
				}
				switch request := event.(type) {
				case *closeMsg:
					if len(c.requests) == 0 {
						close(c.eventLoopCh)
						return
					}

				case *loopEvent:
					request.sequence = c.seq
					c.seq++
					request.out, request.action = c.conn.OnData(request.in)
					c.dataEvents = append(c.dataEvents, request)
					c.afterLoopRun(request)
					if c.closed {
						close(c.eventLoopCh)
						return
					}

				case *requestGroup:
					if len(request.requests) > 0 {

						var in []byte
						for _, r := range request.requests {
							in = append(in, r.request...)
							c.requests = append(c.requests, r)
						}

						c.queue = append(c.queue, request)

						out, action := c.conn.OnData(in)

						event := &loopEvent{
							in:     in,
							out:    out,
							action: action,
							requests: request.requests,
						}
						event.sequence = c.seq
						c.seq++

						c.dataEvents = append(c.dataEvents, event)

						c.afterLoopRun(event)

						if c.closed {
							close(c.eventLoopCh)
							return
						}
					}
				}
			}
		}
	}()

	// recv
	//c.wg.Add(1)
	//go func() {
	//	defer c.wg.Done()
	//
	//	for {
	//		select {}
	//	}
	//}()
}

func (c *mockEvConn) afterLoopRun(data *loopEvent) {
	c.reader = api.NewReplyReader(data.out)
	//c.reader.Reset(data.out)

	var (
		reply api.CommandReply
		err   error
	)

	if len(data.out) > 0 {
		reply, err = c.reader.ReadReply()

		for err == nil {
			data.replies = append(data.replies, reply)

			if len(c.requests) == 0 {
				// Unexpected reply
				fmt.Println("Unexpected reply")

			} else {
				request := c.requests[0]
				c.requests = c.requests[1:]
				if request != nil {

					request.chResponse <- reply
				}
			}

			reply, err = c.reader.ReadReply()
		}
	}

	if err != nil && err != io.EOF {
		panic(err)
	}

	// Parse responses
	fmt.Println(data.String())
}

func (c *mockEvConn) close() {
	c.eventLoopCh <- &closeMsg{}
	c.cancel()
	c.wg.Wait()
}

func (c *mockEvConn) printData() {

}

// Wakes the connection up if necessary.
func (c *mockEvConn) Wake() error {
	c.eventLoopCh <- &loopEvent{}
	return nil
}

// Context returns a user-defined context.
func (c *mockEvConn) Context() interface{} {
	return nil
}

// SetContext sets a user-defined context.
func (c *mockEvConn) SetContext(interface{}) {

}

// AddrIndex is the index of server address that was passed to the Serve call.
func (c *mockEvConn) AddrIndex() int {
	return 0
}

// LocalAddr is the connection's local socket address.
func (c *mockEvConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is the connection's remote peer address.
func (c *mockEvConn) RemoteAddr() net.Addr {
	return nil
}
