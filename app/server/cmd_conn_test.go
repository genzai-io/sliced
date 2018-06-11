package server_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/genzai-io/sliced/app/cmd"
)

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
		expectedReplies[i] = OfBulk
	}
	conn.SendPacket(beforeList...)

	// Force into worker state
	conn.SendPacket(&cmd.Sleep{})

	// Send a series of single command packets
	for i := 0; i < backlogged; i++ {
		conn.SendPacket(&cmd.Get{Key: fmt.Sprintf("backlogged: %d", i)})
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
	conn := newMockConn()

	conn.SendPacket(&cmd.Get{Key: "hi"})

	conn.SendPacket(&WorkerSleep{})
	conn.SendPacket(&WorkerSleep{})
	conn.SendPacket(&cmd.Get{Key: "hi"})
	conn.SendPacket(&cmd.Get{Key: "hi"})
	conn.SendPacket(&cmd.Get{Key: "hi"})
	conn.SendPacket(&cmd.Get{Key: "hi"})

	packets, err := conn.WaitForPackets(0, conn.requestCount+1, time.Second*3)
	if err != nil {
		t.Fatal(err)
	}

	Dump(packets...)

	replies := RepliesOf(packets)
	if len(replies) != conn.requestCount {
		t.Errorf("Expected the last packet to have %d replies instead of %d replies", conn.requestCount, len(replies))
	}

	AssertReplies(t, replies,
		OfBulk,
		IsOK,
		IsOK,
		OfBulk,
		OfBulk,
		IsBulk("key: hi"),
		IsBulk("key: hi"),
	)

	conn.close()
}
