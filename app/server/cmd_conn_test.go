package server_test

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/cmd"
)

type commands struct {
	name string
}

var emptyCommands = &[]commands{}
var emptyClear = unsafe.Pointer(&emptyCommands)

type connection struct {
	cm *[]commands
}

func TestConn_Ptr(t *testing.T) {
	conn := &connection{
		cm: emptyCommands,
	}

	//next := &[]commands{}
	n := []commands{}
	n = append(n, commands{"1"})

	//atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&conn.cm)), unsafe.Pointer(&n))

	prev := (atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&conn.cm)),
		unsafe.Pointer(&n)))

	fmt.Println(*(*[]commands)(prev))

	prev = (atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&conn.cm)),
		unsafe.Pointer(&emptyCommands)))

	fmt.Println(conn.cm)
}

func TestConn_ParseFloat(t *testing.T) {
	f := float64(1000.1)
	i := int64(f)

	cast := *(*int64)(unsafe.Pointer(&f))
	uncasted := *(*float64)(unsafe.Pointer(&cast))

	value := strconv.FormatFloat(f, 'f', 6, 64)

	strconv.ParseFloat("1000.1", 64)

	fmt.Println(f)
	fmt.Println(cast)
	fmt.Println(uncasted)
	fmt.Println(i)
	fmt.Println(value)
}

func TestCmdConn_Worker(t *testing.T) {
	conn := newMockConn()
	defer conn.close()

	before := 50
	backlogged := 50
	commandCount := before + backlogged + 1

	// 1 before with replies
	// 1 worker queued
	// 50 backlogged queued
	// 1 wake with 51 replies
	expectedPackets := 1 + backlogged + 1 + 1

	expectedReplies := make([]ReplyMatcher, commandCount)

	// Send a single before list
	beforeList := make([]Command, before)
	for i := 0; i < before; i++ {
		beforeList[i] = &cmd.Get{Key: fmt.Sprintf("before: %d", i)}
		expectedReplies[i] = WithBulk
	}
	conn.Send(beforeList...)

	// Force into worker state
	conn.Send(&cmd.Sleep{})

	// Send a series of single command packets
	for i := 0; i < backlogged; i++ {
		conn.Send(&cmd.Get{Key: fmt.Sprintf("backlogged: %d", i)})
	}

	// Wait for all packets
	packets, err := conn.WaitForPackets(0, expectedPackets, time.Second*3)
	if err != nil {
		t.Fatal(err)
	}

	//Dump(packets...)

	// First packet should have a reply
	if len(packets[0].replies) != before {
		t.Errorf("first packet expected %d replies instead of %d", before, len(packets[0].replies))
	}
	lastPacket := packets[len(packets)-1]
	if len(lastPacket.replies) != backlogged+1 {
		t.Errorf("last packet expected %d replies instead of %d", backlogged+1, len(lastPacket.replies))
	}
	for i := 1; i < len(packets)-2; i++ {
		if len(packets[i].replies) > 0 {
			t.Errorf("expected no replies on packet: \n%s", packets[i].String())
		}
	}
	replies := RepliesOf(packets)
	if len(replies) != commandCount {
		t.Errorf("Expected the last packet to have %d replies instead of %d replies", commandCount, len(replies))
	}
}

//
func TestCmdConn_MultipleWorkers(t *testing.T) {
	defaultTimeout = time.Hour

	conn := newMockConn()
	defer conn.close()

	conn.
		Send(&cmd.Get{Key: "hi"}).
		ShouldReply(t, WithBulk).

		Send(&cmd.Sleep{Millis: 5000}).
		ShouldNotReply(t).

		Send(&cmd.Sleep{Millis: 50}).
		ShouldNotReply(t).

		Send(&cmd.Get{Key: "hi"}).
		ShouldNotReply(t).

		Send(&cmd.Get{Key: "hi"}).
		ShouldNotReply(t).

		Send(&cmd.Get{Key: "hi"}).
		ShouldNotReply(t).

		Send(&cmd.Get{Key: "hi"}).
		ShouldNotReply(t).conn.

		ShouldWakeWithin(t, defaultTimeout).
		ShouldReply(t,
		WithOK,
		WithOK,
		WithBulk,
		WithBulk,
		BulkValue("key: hi"),
		BulkValue("key: hi")).conn.

		ShouldCountReplies(t, len(conn.Commands())).
		DumpPackets()
}

//
func TestCmdConn_MultipleWorkers2(t *testing.T) {
	defaultTimeout = time.Hour
	dumpPackets = false

	conn := newMockConn()
	defer conn.close()

	for i := 0; i < 2; i++ {
		conn.
			Send(&cmd.Get{Key: "hi"}).
			//ShouldReply(t, WithBulk).

			Send(&cmd.Sleep{Millis: 50}).
			//ShouldNotReply(t).
		//
			Send(&cmd.Sleep{Millis: 50}).
			//ShouldNotReply(t).
		//
			Send(&cmd.Get{Key: "hi"}).
			//ShouldNotReply(t).
		//
			Send(&cmd.Get{Key: "hi"})
			//ShouldNotReply(t).conn.
		//
		//Send(&cmd.Get{Key: "hi"}).
		//ShouldNotReply(t).
		//
		//Send(&cmd.Get{Key: "hi"}).
		//ShouldNotReply(t).conn.
		//

		//ShouldReply(t,
		//WithOK,
		//WithOK,
		//WithBulk,
		//WithBulk,
		//BulkValue("key: hi"),
		//BulkValue("key: hi")).conn.
		//
		//ShouldCountReplies(t, len(conn.Commands())).
		//DumpPackets()

			//ShouldWakeWithin(t, defaultTimeout)


	}

	//conn.DumpPackets()

	time.Sleep(time.Second)

	conn.DumpPackets()

	//conn.ShouldWakeWithin(t, time.Second * 2)

	//Dump(conn.Packets()...)
}

//
func TestCmdConn_Multi(t *testing.T) {
	conn := newMockConn()
	defer conn.close()

	conn.
		Send(api.BulkString("multi")).
		ShouldReply(t, WithOK).

		Send(&cmd.Get{Key: "a"}).
		ShouldReply(t, IsQueued).

		Send(&cmd.Get{Key: "a"}).
		ShouldReply(t, IsQueued).

		Send(api.BulkString("exec")).
		//ShouldReply(t, WithOK).

		conn.

		DumpPackets()
}
