package server

import (
	"syscall"
	"unsafe"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/fastlane"
)

type CmdClient interface {
	SendAsync(command api.Command)

	SendMultiAsync()
}

type localCmdClient struct {

}

type CmdClientSocket struct {
	ev     evio.Conn
	target syscall.Sockaddr
}

func (c *CmdClientSocket) send() {

}

// Channel of *cmdGroup
type proxyWorkerChan struct {
	base fastlane.ChanPointer
}

func (ch *proxyWorkerChan) Send(value *proxyRequest) {
	ch.base.Send(unsafe.Pointer(value))
}

func (ch *proxyWorkerChan) Recv() *cmdGroup {
	return (*cmdGroup)(ch.base.Recv())
}

type proxyRequest struct {
	command api.Command
	reply   api.Command

	onReply func(request *proxyRequest)
}


type proxyCommand interface {
	OnReply(reply api.CommandReply) bool
}

//
type Multi struct {
	requests []api.Command
	replies  []api.CommandReply
}
