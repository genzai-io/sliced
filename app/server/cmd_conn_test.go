package server

import (
	"github.com/genzai-io/sliced/common/evio"
	"net"
	"testing"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/app/api"
	"sync"
	"context"
	"fmt"
	"strings"
	"io"
	"time"
)

type WorkerSleep = cmd.Sleep

func TestConn_OnData(t *testing.T) {
	conn := newMockConn()

	backlogged := 50
	commandCount := backlogged + 1
	expectedPackets := commandCount + 1

	//conn.send(&cmd.Get{Key: "hi"})

	conn.send(&WorkerSleep{})
	for i := 0; i < backlogged; i++ {
		conn.send(&cmd.Get{Key: fmt.Sprintf("%d", i)})
	}

	packets := conn.waitForPackets(0, expectedPackets)

	PrintAllEvents(packets...)

	// First packet should have a reply
	//if len(packets[0].replies) != 1 {
	//	t.Errorf("First command expected reply")
	//}
	for i := 1; i < backlogged; i++ {
		if len(packets[i].replies) > 0 {
			t.Errorf("Expected no replies")
		}
	}
	replies := gatherReplies(packets)
	if len(replies) != commandCount {
		//packet := conn.waitForPackets(expectedPackets, 1)
		//_ = packet
		t.Errorf("Expected the last packet to have %d replies instead of %d replies", commandCount, len(replies))
	}

	conn.close()
}

func TestConn_OnData2(t *testing.T) {
	conn := newMockConn()

	//commandCount := backlogged + 1
	//expectedPackets := commandCount + 1

	conn.send(&cmd.Get{Key: "hi"})

	conn.send(&WorkerSleep{})
	conn.send(&WorkerSleep{})
	conn.send(&cmd.Get{Key: "hi"})
	conn.send(&cmd.Get{Key: "hi"})
	conn.send(&cmd.Get{Key: "hi"})
	conn.send(&cmd.Get{Key: "hi"})

	packets := conn.waitForPackets(0, conn.sendCount+1)

	PrintAllEvents(packets...)

	replies := gatherReplies(packets)
	if len(replies) != conn.sendCount {
		//packet := conn.waitForPackets(expectedPackets, 1)
		//_ = packet
		t.Errorf("Expected the last packet to have %d replies instead of %d replies", conn.sendCount, len(replies))
	}

	assertReplies(t, replies,
		api.String("key: hi"),
		api.OK,
		api.OK,
		api.String("key: hi"),
		api.String("key: hi"),
		api.String("key: hi"),
		api.String("key: hi"),
	)

	//for _, reply := range replies {
	//	switch t := reply.(type) {
	//	case api.String:
	//		fmt.Println("String")
	//
	//	case api.BulkString:
	//		fmt.Println("BulkString")
	//
	//	case api.Bytes:
	//		fmt.Println("Bytes")
	//
	//	case api.Ok:
	//		fmt.Println("Ok")
	//
	//	case api.Queued:
	//		fmt.Println("Queued")
	//	}
	//}
	//fmt.Println(len(replies))

	conn.close()
}

func newMockConn() *mockEvConn {
	ctx, cancel := context.WithCancel(context.Background())
	ev := &mockEvConn{
		ctx:         ctx,
		cancel:      cancel,
		eventLoopCh: make(chan interface{}, 10000),
		reader:      api.NewReplyReader([]byte{}),
		dataCh:      make(chan *loopEvent, 10000),
		replyCh:     make(chan api.CommandReply, 10000),
		//sendCh:      make(chan *requestGroup, 1),
		//recvCh:      make(chan *command, 1),
	}
	conn := newConn(ev)
	ev.conn = conn
	ev.onData = conn.OnData

	ev.start()
	return ev
}

type loopEvent struct {
	sequence int
	in       []byte
	out      []byte
	action   evio.Action

	pending  int
	requests []*command
	replies  []api.CommandReply
}

func (l *loopEvent) IsQueued() bool {
	return l.pending+len(l.requests) > len(l.replies)
}

func (l *loopEvent) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\nEvent: %d\n", l.sequence+1))

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

			if l.IsQueued() {
				builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests)-len(l.replies)))
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
			builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests)-len(l.replies)))
			builder.WriteString("===============================\n\n")
		} else {
			builder.WriteString(fmt.Sprintf("%d RESPONSE(s) [%d bytes]\n", len(l.replies), len(l.out)))
			builder.WriteString("-------------------------------\n")
			builder.WriteString(string(l.out))

			if l.IsQueued() {
				builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests)-len(l.replies)))
			}
			builder.WriteString("===============================\n\n")
		}
	}

	return builder.String()
}

func assertReplies(t *testing.T, replies []api.CommandReply, shouldMatch ...api.CommandReply) {
	if len(replies) != len(shouldMatch) {
		t.Errorf("received %d, but expected %d", len(replies), len(shouldMatch))
	}
	for index, reply := range replies {
		match := shouldMatch[index]

		if !api.ReplyEquals(reply, match) {
			t.Errorf("replies[%d] %s != %s", index, reply, match)
		}
	}
}

func PrintAllEvents(events ...*loopEvent) {
	for _, event := range events {
		fmt.Println(event.String())
	}
}

func gatherReplies(events []*loopEvent) (out []api.CommandReply) {
	for _, event := range events {
		out = append(out, event.replies...)
	}
	return
}

type requestGroup struct {
	requests []*command
	in       []byte
	out      []byte

	replyCounter int

	chResponse chan interface{}
}

//func (r *requestGroup) wait() *requestGroup {
//	if len(r.requests) == 0 {
//		return r
//	}
//	for _, request := range r.requests {
//		request.response()
//		r.replyCounter++
//	}
//	close(r.chResponse)
//	return r
//}

type command struct {
	command api.Command
	request []byte
	reply   api.CommandReply
	replies []api.CommandReply

	chResponse chan api.CommandReply
}

//func (r *command) response() *command {
//	select {
//	case reply, ok := <-r.chResponse:
//		if !ok {
//			return r
//		}
//		r.reply = reply
//		close(r.chResponse)
//		return r
//
//	case <-time.After(10 * time.Second):
//		close(r.chResponse)
//		return r
//	}
//	return r
//}

func (r *command) addReply(reply api.CommandReply) {
	if reply == api.QUEUED {

	}
}

type mockEvConn struct {
	evio.Conn

	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex
	wg     sync.WaitGroup
	conn   *Conn
	onData func(in []byte) (out []byte, action evio.Action)

	closed bool
	seq    int
	queue  []*requestGroup

	loopEvents []*loopEvent
	replies    []api.CommandReply

	replyIndex int
	pending    []*command

	eventLoopCh chan interface{}

	dataCh  chan *loopEvent
	replyCh chan api.CommandReply
	//sendCh chan *requestGroup
	//recvCh chan *command

	reader *api.ReplyReader

	sendCount int
}

func (c *mockEvConn) beginMulti(commands ...api.Command) {

}

func (c *mockEvConn) send(commands ...api.Command) *requestGroup {
	group := &requestGroup{
		chResponse: make(chan interface{}, 1),
	}
	c.sendCount += len(commands)
	for _, cm := range commands {
		request := &command{
			command:    cm,
			request:    cm.Marshal(nil),
			chResponse: make(chan api.CommandReply, 1),
		}
		group.requests = append(group.requests, request)
		group.in = append(group.in, request.request...)
	}

	c.eventLoopCh <- group
	return group
}

func (c *mockEvConn) waitForPackets(begin, count int) []*loopEvent {
	c.Lock()
	if len(c.loopEvents) >= begin+count {
		s := c.loopEvents[begin : begin+count]
		c.Unlock()
		return s
	}
	c.Unlock()
	for {
		select {
		case <-c.dataCh:
			c.Lock()
			if len(c.loopEvents) >= begin+count {
				s := c.loopEvents[begin : begin+count]
				c.Unlock()
				return s
			}
			c.Unlock()

		case <-time.After(time.Second * 5):
			return nil
		}
	}
}

func (c *mockEvConn) waitForReplies(begin, count int) []api.CommandReply {
	c.Lock()
	if len(c.replies) >= begin+count {
		s := c.replies[begin : begin+count]
		c.Unlock()
		return s
	}
	c.Unlock()
	for {
		select {
		case <-c.replyCh:
			c.Lock()
			if len(c.replies) >= begin+count {
				s := c.replies[begin : begin+count]
				c.Unlock()
				return s
			}
			c.Unlock()

		case <-time.After(time.Second * 5):
			return nil
		}
	}
}

type closeMsg struct{}

func (c *mockEvConn) start() {
	// event loop
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

	Loop:
		for {
			select {

			case <-c.ctx.Done():
				return

			case event, ok := <-c.eventLoopCh:
				if !ok {
					return
				}
				switch msg := event.(type) {
				case *closeMsg:
					if len(c.pending) == 0 {
						close(c.eventLoopCh)
						return
					}

				case *loopEvent:
					// Set sequence
					msg.sequence = c.seq
					c.seq++
					// event-loop pass
					msg.out, msg.action = c.conn.OnData(msg.in)

					c.afterLoopRun(msg)

					if c.closed {
						close(c.eventLoopCh)
						return
					}

				case *requestGroup:
					if len(msg.requests) == 0 {
						continue Loop
					}

					var in []byte
					for _, r := range msg.requests {
						in = append(in, r.request...)
					}

					c.queue = append(c.queue, msg)

					// event-loop pass
					out, action := c.conn.OnData(in)

					// create a loop event
					event := &loopEvent{
						sequence: c.seq,
						in:       in,
						out:      out,
						action:   action,
						requests: msg.requests,
					}
					c.seq++

					// handle loop event
					c.afterLoopRun(event)

					if c.closed {
						close(c.eventLoopCh)
						return
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
	// Create a reply reader
	c.reader = api.NewReplyReader(data.out)
	//c.reader.Reset(data.out)

	var (
		reply api.CommandReply
		err   error
	)

	if len(data.out) > 0 {
		reply, err = c.reader.Next()

		for err == nil {
			data.replies = append(data.replies, reply)

			c.Lock()
			c.replies = append(c.replies, reply)
			c.Unlock()

			c.replyCh <- reply

			reply, err = c.reader.Next()
		}
	}

	c.Lock()
	c.loopEvents = append(c.loopEvents, data)
	c.Unlock()

	c.dataCh <- data

	if err != nil && err != io.EOF {
		panic(err)
	}

	// Parse responses
	//fmt.Println(data.String())
}

func (c *mockEvConn) close() {
	c.eventLoopCh <- &closeMsg{}
	c.cancel()
	c.wg.Wait()
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
