// Testing Conn requires simulating the event-loop. That's what the structures
// in this file do. It mocks an evio.Conn and the evio event-loop. Then, many high
// level helpers, a network packet abstraction and a RESP protocol reply parser to
// parse the network packets as a series of RESP replies.
//
//
package cmd_test

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
	cmd_server "github.com/genzai-io/sliced/app/server/cmd"
	"github.com/genzai-io/sliced/common/evio"
)

var (
	ErrTimeout = errors.New("timeout")

	defaultTimeout = time.Second * 5

	dumpPackets = false
)

type Command = api.Command
type Reply = api.CommandReply

func AssertReplies(t *testing.T, replies []Reply, shouldMatch ...ReplyMatcher) {
	if len(replies) != len(shouldMatch) {
		t.Errorf("received %d, but expected %d", len(replies), len(shouldMatch))
	}
	if len(replies) == 0 {
		return
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
		eventLoopCh: make(chan interface{}, 1000),
		dataCh:      make(chan *packet, 1000),
		replyCh:     make(chan api.CommandReply, 1000),
		wakeCh:      make(chan *packet, 1000),
	}
	conn := cmd_server.NewConn(ev)
	ev.conn = conn

	ev.start()
	return ev
}

// Mocks an event loop data event
type packet struct {
	sequence int
	conn     *mockEvConn

	// Conn data
	in     []byte
	out    []byte
	action evio.Action

	// State
	beforeRequests int
	beforeReplies  int
	afterReplies   int

	pending int

	// Requests
	requests []Command
	// Replies
	replies []Reply

	ch chan *packet
}

func (l *packet) Wait() *packet {
	select {
	case p, _ := <-l.ch:
		return p

	case <-time.After(time.Second * 120):
		panic("timed-out waiting for packet")
	}
	return l
}

func (l *packet) Send(command ...Command) *packet {
	return l.conn.Send(command...)
}

func (l *packet) IsWake() bool {
	return len(l.in) == 0
}

func (l *packet) IsScheduled() bool {
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

			if l.IsScheduled() {
				builder.WriteString(fmt.Sprintf("%d SCHEDULED\n", len(l.requests)-len(l.replies)))
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
			builder.WriteString(fmt.Sprintf("%d SCHEDULED\n", len(l.requests)-len(l.replies)))
			builder.WriteString("===============================\n\n")
		} else {
			builder.WriteString(fmt.Sprintf("%d RESPONSE(s) [%d bytes]\n", len(l.replies), len(l.out)))
			builder.WriteString("-------------------------------\n")
			builder.WriteString(string(l.out))

			if l.IsScheduled() {
				builder.WriteString(fmt.Sprintf("%d QUEUED\n", len(l.requests)-len(l.replies)))
			}
			builder.WriteString("===============================\n\n")
		}
	}

	return builder.String()
}

func (l *packet) ExpectOK(t *testing.T) *packet {
	return l.ShouldReply(t, WithOK)
}

func (l *packet) ExpectQueued(t *testing.T) *packet {
	return l.ShouldReply(t, WithQueued)
}

func (l *packet) ExpectBulk(t *testing.T) *packet {
	return l.ShouldReply(t, WithBulk)
}

func (l *packet) ExpectArray(t *testing.T) *packet {
	return l.ShouldReply(t, WithArray)
}

func (l *packet) ExpectError(t *testing.T) *packet {
	return l.ShouldReply(t, WithError)
}

func (l *packet) ExpectEmpty(t *testing.T) *packet {
	return l.ShouldNotReply(t)
}

func (l *packet) ShouldReply(t *testing.T, matchers ...ReplyMatcher) *packet {
	AssertReplies(t, l.replies, matchers...)
	return l
}

func (l *packet) ShouldNotReply(t *testing.T) *packet {
	if len(l.replies) > 0 {
		t.Errorf("expected no replies instead received %d", len(l.replies))
	}
	return l
}

// Dumps a human readable console representation of each simulated network packet.
func Dump(events ...*packet) {
	for _, event := range events {
		fmt.Println(event.String())
	}
}

// Gathers all replies in order from all supplied packets
func GatherReplies(events []*packet) (out []api.CommandReply) {
	for _, event := range events {
		out = append(out, event.replies...)
	}
	return
}

// A mock event-loop connection.
// It mocks the event-loop by simulating it on a channel and feeding it "packets".
type mockEvConn struct {
	evio.Conn

	ctx    context.Context
	cancel context.CancelFunc
	sync.Mutex
	wg     sync.WaitGroup
	conn   *cmd_server.Conn

	closed bool
	seq    int

	packets      []*packet
	packetIndex  int
	commands     []Command
	commandCount int
	replies      []Reply
	replyIndex   int

	wakeIndex int
	wakes     []*packet

	eventLoopCh chan interface{}

	dataCh  chan *packet
	replyCh chan api.CommandReply
	wakeCh  chan *packet
}

// Send "x" commands as a single network "packet".
// It will wait for the packet to go through a single
// event-loop pass. When this returns, the packet may
// or may not have any data "Out" data with RESP replies
// automatically parsed.
func (c *mockEvConn) Send(commands ...Command) *packet {
	c.Lock()

	c.commandCount += len(commands)

	packet := &packet{
		sequence: c.seq,
		conn:     c,
		ch:       make(chan *packet, 1),
	}
	c.seq++

	packet.requests = append(packet.requests, commands...)
	c.commands = append(c.commands, commands...)
	for _, cm := range commands {
		packet.in = cm.Marshal(packet.in)
	}

	c.Unlock()

	select {
	case c.eventLoopCh <- packet:
	default:
		panic("send eventLoopCh failed")
	}

	packet.Wait()
	return packet
}

func (c *mockEvConn) Packets() []*packet {
	c.Lock()
	defer c.Unlock()

	var packets []*packet
	return append(packets, c.packets...)
}

func (c *mockEvConn) Wakes() []*packet {
	c.Lock()
	defer c.Unlock()

	var packets []*packet
	return append(packets, c.wakes...)
}

func (c *mockEvConn) Replies() []Reply {
	c.Lock()
	defer c.Unlock()

	var replies []Reply
	return append(replies, c.replies...)
}

func (c *mockEvConn) Commands() []Command {
	c.Lock()
	defer c.Unlock()

	var commands []Command
	return append(commands, c.commands...)
}

func (c *mockEvConn) AllReplies(t *testing.T) ([]Reply, error) {
	replies, err := c.WaitForReplies(0, c.commandCount, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return replies, err
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

func (c *mockEvConn) WaitForReplies(begin, count int, timeout time.Duration) ([]Reply, error) {
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

func (c *mockEvConn) WaitForWakePackets(begin, count int, timeout time.Duration) ([]*packet, error) {
	c.Lock()
	if len(c.wakes) >= begin+count {
		s := c.wakes[begin : begin+count]
		c.Unlock()
		return s, nil
	}
	c.Unlock()
	for {
		select {
		case <-c.dataCh:
			c.Lock()
			if len(c.wakes) >= begin+count {
				s := c.wakes[begin : begin+count]
				c.Unlock()
				return s, nil
			}
			c.Unlock()

		case <-time.After(timeout):
			var r []*packet
			c.Lock()
			l := len(c.wakes)
			if l > 0 {
				r = c.wakes[:l]
			}
			c.Unlock()
			return r, ErrTimeout
		}
	}
}

func (c *mockEvConn) Packet(t *testing.T) *packet {
	packets, err := c.WaitForPackets(c.packetIndex, 1, defaultTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) == 0 {
		return nil
	} else {
		return packets[0]
	}
}

func (c *mockEvConn) Expect(t *testing.T, matchers ...ReplyMatcher) {
	c.ExpectTimeout(t, defaultTimeout, matchers...)
}

func (c *mockEvConn) ExpectTimeout(t *testing.T, timeout time.Duration, matchers ...ReplyMatcher) {
	replies, err := c.WaitForReplies(c.replyIndex, len(matchers), timeout)
	c.replyIndex += len(replies)
	if err != nil {
		t.Fatal(err)
	}
	AssertReplies(t, replies, matchers...)
}

func (c *mockEvConn) ShouldWakeWithin(t *testing.T, timeout time.Duration) *packet {
	packets, err := c.WaitForWakePackets(c.wakeIndex, 1, timeout)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) == 0 {
		return nil
	} else {
		c.wakeIndex++
		return packets[0]
	}
}

func (c *mockEvConn) ShouldCountReplies(t *testing.T, numberOfReplies int) *mockEvConn {
	c.Lock()
	count := len(c.replies)
	c.Unlock()

	if count != numberOfReplies {
		t.Errorf("expect reply count %d does not match actual %d", numberOfReplies, count)
	}
	return c
}

func (c *mockEvConn) DumpPackets() {
	Dump(c.Packets()...)
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
				switch msg := event.(type) {
				case *closeMsg:
					//if len(c.pending) == 0 {
					c.closed = true
					//close(c.eventLoopCh)
					return
					//}

				case *packet:
					if c.closed {
						c.finishLoop(msg)
						return
					}

					// event-loop pass
					msg.out, msg.action = c.conn.OnData(msg.in)

					// finish
					c.finishLoop(msg)
				}
			}
		}
	}()
}

func (c *mockEvConn) finishLoop(p *packet) {
	p.conn = c

	// Let's read the replies and finish up the event-loop pass
	reader := api.NewReplyReader(p.out)

	var (
		reply api.CommandReply
		err   error
	)

	if len(p.out) > 0 {
		reply, err = reader.Next()

		for err == nil {
			// Add reply to packet
			p.replies = append(p.replies, reply)

			// Add reply to the connection
			//c.Lock()
			//c.replies = append(c.replies, reply)
			//c.Unlock()

			// Notify of new reply
			//c.replyCh <- reply

			reply, err = reader.Next()
		}
	}

	// Add packet to connection
	c.Lock()
	c.replies = append(c.replies, p.replies...)
	if len(p.in) == 0 {
		c.wakes = append(c.wakes, p)
	}
	c.packets = append(c.packets, p)
	c.Unlock()

	// Notify of new packet
	c.dataCh <- p

	if len(p.in) == 0 {
		c.wakeCh <- p
	}

	p.ch <- p

	for _, reply := range p.replies {
		c.replyCh <- reply
	}

	if dumpPackets {
		Dump(p)
	}

	if err != nil && err != io.EOF {
		panic(err)
	}
}

func (c *mockEvConn) close() {
	c.eventLoopCh <- &closeMsg{}
	c.cancel()
	close(c.eventLoopCh)
	//c.wg.Wait()
}

///////////////////////////////////////////////////////
// evio.Conn Interface Implementation
///////////////////////////////////////////////////////

// Wakes the connection up if necessary.
func (c *mockEvConn) Wake() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	c.Lock()
	seq := c.seq
	c.seq++
	c.Unlock()
	c.eventLoopCh <- &packet{conn: c, sequence: seq, ch: make(chan *packet, 1)}
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

///////////////////////////////////////////////////////
// api.CommandReply assertion
///////////////////////////////////////////////////////

type ReplyMatcher interface {
	Test(reply Reply) bool

	String() string
}

var (
	WithSimpleString = ofSimpleString{}
	WithBulk         = ofBulkString{}
	WithInt          = ofInt{}
	WithArray        = ofArray{}
	WithOK           = isOK{}
	WithQueued       = isQueued{}
	WithError        = ofErr{}
)

//
type ofErr struct{}

func (i ofErr) Test(reply Reply) bool {
	switch reply.(type) {
	case api.Err:
		return true
	}
	return false
}
func (i ofErr) String() string {
	return "WithError"
}

func IsErr(value string) ReplyMatcher {
	return isSimpleString{value}
}

type isErr struct {
	value string
}

func (i isErr) Test(reply Reply) bool {
	switch v := reply.(type) {
	case api.Err:
		return string(v) == i.value
	}
	return false
}
func (i isErr) String() string {
	return fmt.Sprintf("Err(\"%s\")", i.value)
}

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
	return "WithSimpleString"
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
	return "WithBulk"
}

func BulkValue(value string) ReplyMatcher {
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
	return fmt.Sprintf("Int(%d)", i.value)
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

func ArrayValue(value api.Array) ReplyMatcher {
	return ArrayValue(value)
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
