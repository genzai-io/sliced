package server

import (
	"sync"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/evio"
)

//
type NodeConn struct {
	sync.Mutex

	evc    evio.Conn
	kind   api.ConnKind
	action evio.Action // event loop status
	done   bool        // flag to signal it's done

	out []byte

	clientc *Conn
	backlog []api.Command
}

func (n *NodeConn) Send(command api.Command) {
	n.Lock()
	if len(n.backlog) > 0 {
		n.backlog = append(n.backlog, command)
	} else {

	}
	n.Unlock()
}

func (n *NodeConn) wake() {
	n.evc.Wake()
}

func (n *NodeConn) onLoop(in []byte) ([]byte, evio.Action) {
	if in == nil {

	}

	n.Lock()

	n.Unlock()

	return n.out, n.action
}