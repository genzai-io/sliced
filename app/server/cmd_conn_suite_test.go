package server_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/app/server"
	"github.com/genzai-io/sliced/common/evio"
)

var (
	ErrTimeout = errors.New("timeout")
)

type WorkerSleep = cmd.Sleep
type Command = api.Command
type Reply = api.CommandReply

func AssertReplies(t *testing.T, replies []Reply, shouldMatch ...ReplyMatcher) {
	if len(replies) != len(shouldMatch) {
		t.Errorf("received %d, but expected %d", len(replies), len(shouldMatch))
	}
	for index, reply := range replies {
		match := shouldMatch[index]

		if !match.Test(reply) {
			t.Errorf("replies[%d] %s != %s", index, reply, match)
		}
	}
}

func newMockConn() *mockEvConn {
	ctx, cancel := context.WithCancel(context.Background())
	ev := &mockEvConn{
		ctx:         ctx,
		cancel:      cancel,
		reader:      api.NewReplyReader([]byte{}),
		eventLoopCh: make(chan interface{}, 10000),
		dataCh:      make(chan *packet, 10000),
		replyCh:     make(chan api.CommandReply, 10000),
	}
	conn := server.NewConn(ev)
	ev.conn = conn
	ev.onData = conn.OnData

	ev.start()
	return ev
}

type packet struct {
	sequence int
	in       []byte
	out      []byte
	action   evio.Action

	pending  int
	requests []*command
	replies  []api.CommandReply
}

func (l *packet) IsQueued() bool {
	return l.pending+len(l.requests) > len(l.replies)
}

func (l *packet) String() string {
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

func Dump(events ...*packet) {
	for _, event := range events {
		fmt.Println(event.String())
	}
}

func RepliesOf(events []*packet) (out []api.CommandReply) {
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

type command struct {
	command api.Command
	request []byte
	reply   api.CommandReply
	replies []api.CommandReply

	chResponse chan api.CommandReply
}

type mockEvConn struct {
	evio.Conn

	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex
	wg     sync.WaitGroup
	conn   *server.CmdConn
	onData func(in []byte) (out []byte, action evio.Action)

	closed bool
	seq    int
	queue  []*requestGroup

	packets []*packet
	replies []Reply

	replyIndex int
	pending    []*command

	eventLoopCh chan interface{}

	dataCh  chan *packet
	replyCh chan api.CommandReply
	//sendCh chan *requestGroup
	//recvCh chan *command

	reader *api.ReplyReader

	requestCount int
}

func (c *mockEvConn) SendPacket(commands ...Command) *requestGroup {
	group := &requestGroup{
		chResponse: make(chan interface{}, 1),
	}
	c.requestCount += len(commands)
	for _, cm := range commands {
		request := &command{
			command: cm,
			request: cm.Marshal(nil),
		}
		group.requests = append(group.requests, request)
		group.in = append(group.in, request.request...)
	}

	c.eventLoopCh <- group
	return group
}

func (c *mockEvConn) WaitForPackets(begin, count int, timeout time.Duration) ([]*packet, error) {
	c.Lock()
	if len(c.packets) >= begin+count {
		s := c.packets[begin : begin+count]
		c.Unlock()
		return s, nil
	}
	c.Unlock()
	for {
		select {
		case <-c.dataCh:
			c.Lock()
			if len(c.packets) >= begin+count {
				s := c.packets[begin : begin+count]
				c.Unlock()
				return s, nil
			}
			c.Unlock()

		case <-time.After(timeout):
			var r []*packet
			c.Lock()
			l := len(c.packets)
			if l > 0 {
				r = c.packets[:l]
			}
			c.Unlock()
			return r, ErrTimeout
		}
	}
}

func (c *mockEvConn) WaitForReplies(begin, count int, timeout time.Duration) ([]api.CommandReply, error) {
	c.Lock()
	if len(c.replies) >= begin+count {
		s := c.replies[begin : begin+count]
		c.Unlock()
		return s, nil
	}
	c.Unlock()
	for {
		select {
		case <-c.replyCh:
			c.Lock()
			if len(c.replies) >= begin+count {
				s := c.replies[begin : begin+count]
				c.Unlock()
				return s, nil
			}
			c.Unlock()

		case <-time.After(timeout):
			var r []Reply
			c.Lock()
			l := len(c.replies)
			if l > 0 {
				r = c.replies[:l]
			}
			c.Unlock()
			return r, ErrTimeout
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

				case *packet:
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
					event := &packet{
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

func (c *mockEvConn) afterLoopRun(data *packet) {
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
	c.packets = append(c.packets, data)
	c.Unlock()

	c.dataCh <- data

	if err != nil && err != io.EOF {
		panic(err)
	}

	// Parse responses
	//fmt.Println(data.IsSimpleString())
}

func (c *mockEvConn) close() {
	c.eventLoopCh <- &closeMsg{}
	c.cancel()
	c.wg.Wait()
}

// Wakes the connection up if necessary.
func (c *mockEvConn) Wake() error {
	c.eventLoopCh <- &packet{}
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

type ReplyMatcher interface {
	Test(reply Reply) bool

	String() string
}

var (
	OfSimpleString = ofSimpleString{}
	OfBulk         = ofBulkString{}
	OfInt          = ofInt{}
	OfArray        = ofArray{}
	IsOK           = isOK{}
	IsQueued       = isQueued{}
)
//
type ofSimpleString struct{}

func (i ofSimpleString) Test(reply Reply) bool {
	switch reply.(type) {
	case api.SimpleString:
		return true
	}
	return false
}
func (i ofSimpleString) String() string {
	return "OfSimpleString"
}

func IsSimpleString(value string) ReplyMatcher {
	return isSimpleString{value}
}

type isSimpleString struct {
	value string
}

func (i isSimpleString) Test(reply Reply) bool {
	switch v := reply.(type) {
	case api.SimpleString:
		return string(v) == i.value
	}
	return false
}
func (i isSimpleString) String() string {
	return fmt.Sprintf("SimpleString(\"%s\")", i.value)
}

//
type ofBulkString struct{}

func (i ofBulkString) Test(reply Reply) bool {
	switch reply.(type) {
	case api.BulkString:
		return true
	case api.Bulk:
		return true
	}
	return false
}
func (i ofBulkString) String() string {
	return "OfBulk"
}

func IsBulk(value string) ReplyMatcher {
	return isBulkString{value}
}

//
type isBulkString struct {
	value string
}

func (i isBulkString) Test(reply Reply) bool {
	switch v := reply.(type) {
	case api.BulkString:
		return i.value == string(v)
	case api.Bulk:
		return i.value == string(v)
	}
	return false
}
func (i isBulkString) String() string {
	return fmt.Sprintf("Bulk(\"%s\")", i.value)
}

//
type isOK struct{}

func (i isOK) Test(reply Reply) bool {
	switch reply.(type) {
	case api.Ok:
		return true
	}
	return false
}
func (i isOK) String() string {
	return "Ok"
}

//
type isQueued struct{}

func (i isQueued) Test(reply Reply) bool {
	switch reply.(type) {
	case api.Queued:
		return true
	}
	return false
}
func (i isQueued) String() string {
	return "Queued"
}

//
type ofInt struct{}

func (i ofInt) Test(reply Reply) bool {
	switch reply.(type) {
	case api.Int:
		return true
	}
	return false
}
func (i ofInt) String() string {
	return "ofInt"
}

func IsInt(value api.Int) ReplyMatcher {
	return IsInt(value)
}

//
type isInt struct {
	value api.Int
}

func (i isInt) Test(reply Reply) bool {
	switch v := reply.(type) {
	case api.Int:
		return i.value == v
	}
	return false
}
func (i isInt) String() string {
	return fmt.Sprintf("Int(%s)", i.value)
}

//
type ofArray struct{}

func (i ofArray) Test(reply Reply) bool {
	switch reply.(type) {
	case api.Array:
		return true
	}
	return false
}
func (i ofArray) String() string {
	return "ofArray"
}

func IsArray(value api.Array) ReplyMatcher {
	return IsArray(value)
}

//
type isArray struct {
	value api.Array
}

func (i isArray) Test(reply Reply) bool {
	switch v := reply.(type) {
	case api.Array:
		if len(i.value) != len(v) {
			return false
		}
		for index, value := range i.value {
			return AssertReply(value, v[index])
		}
	}
	return false
}
func (i isArray) String() string {
	return fmt.Sprintf("Int(%s)", i.value)
}

func AssertReply(reply Reply, match Reply) bool {
	switch rt := reply.(type) {
	case api.SimpleString:
		switch r2t := match.(type) {
		case api.Bulk:
			return string(rt) == string(r2t)
		case api.SimpleString:
			return string(rt) == string(r2t)
			//case ofSimpleString:
			//	return true
		case api.BulkString:
			return string(rt) == string(r2t)
		}
		return false

	case api.BulkString:
		switch r2t := match.(type) {
		case api.Bulk:
			return string(rt) == string(r2t)
		case api.SimpleString:
			return string(rt) == string(r2t)
		case api.BulkString:
			return string(rt) == string(r2t)
		}
		return false

	case api.Bulk:
		switch r2t := match.(type) {
		case api.Bulk:
			return string(rt) == string(r2t)
		case api.SimpleString:
			return string(rt) == string(r2t)
		case api.BulkString:
			return string(rt) == string(r2t)
		}
		return false

	case api.Int:
		switch r2t := match.(type) {
		case api.Int:
			return int64(rt) == int64(r2t)
		}
		return false

	case api.Ok:
		if _, ok := match.(api.Ok); ok {
			return true
		} else {
			return false
		}

	case api.Nil:
		if _, ok := match.(api.Nil); ok {
			return true
		} else {
			return false
		}

	case api.Queued:
		if _, ok := match.(api.Queued); ok {
			return true
		} else {
			return false
		}

	case api.Pong:
		if _, ok := match.(api.Pong); ok {
			return true
		} else {
			return false
		}

	case api.Array:
		if av, ok := match.(api.Array); ok {
			if len(rt) != len(av) {
				return false
			}
			for index, value := range rt {
				return AssertReply(value, av[index])
			}
		} else {
			return false
		}
	}
	return false
}
