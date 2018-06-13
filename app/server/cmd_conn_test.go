package server_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/common/fastlane"
)

func TestFastlane(t *testing.T) {
	lane := fastlane.Chan{}
	lp := &lane

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			msg := lp.Recv()
			if msg == nil {
				fmt.Println("Closing")
				return
			} else {
				time.Sleep(250 * time.Millisecond)
				fmt.Println(msg)
			}
		}
	}()

	for i := 0; i < 100; i++ {
		//lane.Send(fmt.Sprintf("Hi %d", i))
		//time.Sleep(time.Millisecond * 250)
	}

	fmt.Println("Finished sending")
	lane.Send(nil)

	wg.Wait()
}

//
func TestCmdConn_MultipleWorkers(t *testing.T) {
	//defaultTimeout = time.Hour
	dumpPackets = false

	conn := newMockConn()
	defer conn.close()

	conn.
		Send(&cmd.Get{Key: "hi"}).
		ShouldReply(t, WithBulk).

		Send(&cmd.Sleep{Millis: 50}).
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

		ShouldWakeWithin(t, time.Second)

	// Simulate enough time to finish all commands.
	time.Sleep(time.Millisecond * 200)

	// Should received all replies
	conn.ShouldCountReplies(t, len(conn.Commands()))

	AssertReplies(t, conn.Replies(),
		WithBulk,
		WithOK,
		WithOK,
		WithBulk,
		WithBulk,
		WithBulk,
		WithBulk,
	)
}

//
func TestCmdConn_MultipleWorkers2(t *testing.T) {
	//defaultTimeout = time.Hour
	dumpPackets = false

	conn := newMockConn()
	defer conn.close()

	sleepTime := 10
	iterations := 25

	for i := 0; i < iterations; i++ {
		conn.
			Send(&cmd.Get{Key: "hi"})
		conn.
			Send(&cmd.Sleep{Millis: int64(sleepTime)})
	}

	// Simulate enough time to finish all commands.
	time.Sleep(time.Millisecond * time.Duration(sleepTime) * time.Duration(iterations) * 2)

	// Should received all replies
	conn.ShouldCountReplies(t, len(conn.Commands()))

	//conn.DumpPackets()
}

//
func TestCmdConn_Multi(t *testing.T) {
	conn := newMockConn()
	defer conn.close()

	conn.
		Send(api.BulkString("multi")).
		ShouldReply(t, WithOK).

		Send(&cmd.Get{Key: "a"}).
		ShouldReply(t, WithQueued).

		Send(&cmd.Get{Key: "b"}).
		ShouldReply(t, WithQueued).

		Send(api.BulkString("exec")).
		ShouldReply(t, WithArray)
}

//
func TestCmdConn_MultiWorker(t *testing.T) {
	//dumpPackets = true

	conn := newMockConn()
	defer conn.close()

	conn.
		Send(api.BulkString("multi")).
		ShouldReply(t, WithOK).

		Send(&cmd.Get{Key: "a"}).
		ShouldReply(t, WithQueued).

		Send(&cmd.Get{Key: "b"}).
		ShouldReply(t, WithQueued).

		Send(&cmd.Sleep{Millis: 50}).
		ShouldReply(t, WithQueued).

	// "exec" should have empty reply packet
		Send(api.BulkString("exec")).
		ShouldNotReply(t)

	// Simulate enough time to finish all commands.
	time.Sleep(time.Millisecond * 100)

	// Should received all replies
	conn.ShouldCountReplies(t, len(conn.Commands()))
}

//
func TestCmdConn_MultiWorkerDiscard(t *testing.T) {
	//dumpPackets = true

	conn := newMockConn()
	defer conn.close()

	conn.
		Send(api.BulkString("multi")).
		ExpectOK(t).

		Send(&cmd.Get{Key: "a"}).
		ExpectQueued(t).

		Send(&cmd.Get{Key: "b"}).
		ExpectQueued(t).

		Send(&cmd.Sleep{Millis: 50}).
		ExpectQueued(t).

		Send(api.BulkString("discard")).
		ExpectOK(t).

	// "exec" should have empty reply packet
		Send(api.BulkString("exec")).
		ExpectError(t)

	// Simulate enough time to finish all commands.
	//time.Sleep(time.Millisecond * 1000)

	// Should received all replies
	conn.ShouldCountReplies(t, len(conn.Commands()))
}
