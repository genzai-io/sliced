// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build darwin netbsd freebsd openbsd dragonfly linux

package evio

import (
	"errors"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"math/rand"

	"github.com/genzai-io/sliced/common/evio/internal"
	"github.com/kavu/go_reuseport"
)

type conn struct {
	fd         int              // file descriptor
	lnidx      int              // listener index in the server lns list
	loopidx    int              // owner loop
	loopfd     int              // owner loop fd
	loop       *loop            // owner loop
	out        []byte           // write buffer
	sa         syscall.Sockaddr // remote socket address
	reuse      bool             // should reuse input buffer
	opened     bool             // connection opened event fired
	action     Action           // next user action
	ctx        interface{}      // user-defined context
	addrIndex  int
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (c *conn) LoopIndex() int             { return c.loopidx }
func (c *conn) Context() interface{}       { return c.ctx }
func (c *conn) SetContext(ctx interface{}) { c.ctx = ctx }
func (c *conn) AddrIndex() int             { return c.addrIndex }
func (c *conn) LocalAddr() net.Addr        { return c.localAddr }
func (c *conn) RemoteAddr() net.Addr       { return c.remoteAddr }
func (c *conn) Action() Action             { return c.action }
func (c *conn) SetAction(action Action)    { c.action = action }
func (c *conn) Wake() error {
	return c.loop.poll.Wake(c.fd)
}

func (s *server) Dial(addr string) error {
	sa, err := Resolve(addr)
	if err != nil {
		return err
	}
	return s.DialSockAddr(sa)
}

func (s *server) DialSockAddr(sa syscall.Sockaddr) error {
	var err error
	var fd int
	switch sa.(type) {
	case *syscall.SockaddrUnix:
		fd, err = syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	case *syscall.SockaddrInet4:
		fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	case *syscall.SockaddrInet6:
		fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	}
	if err != nil {
		return err
	}
	err = syscall.Connect(fd, sa)
	if err != nil && err != syscall.EINPROGRESS {
		syscall.Close(fd)
		return err
	}
	// Set socket to Non-Blocking
	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return err
	}

	// Pick a loop
	var l *loop = nil
	var lidx = 0

	if len(s.loops) > 1 {
		switch s.balance {
		case LeastConnections:
			// Pick the loop with the fewest connections
			for i, lp := range s.loops {
				if l == nil {
					l = lp
					lidx = i
				} else if atomic.LoadInt32(&l.count) < atomic.LoadInt32(&lp.count) {
					l = lp
					lidx = i
				}
			}
		case RoundRobin:
			lidx = int(atomic.LoadUintptr(&s.accepted)) % len(s.loops)
			l = s.loops[lidx]

		case Random:
			lidx = rand.Int() % len(s.loops)
			l = s.loops[lidx]

		default:
			l = s.loops[int(atomic.LoadUintptr(&s.accepted))%len(s.loops)]
		}
	} else if len(s.loops) == 1 {
		l = s.loops[0]
	}

	if l == nil {
		syscall.Close(fd)
		return errors.New("no loops")
	}

	// Increment dialed
	atomic.AddUintptr(&s.dialed, 1)
	// Increment connect count
	atomic.AddInt32(&l.count, 1)

	// Init conn
	c := &conn{fd: fd, sa: sa, lnidx: lidx, loop: l}

	// Put conn into loop map
	l.fdconns[c.fd] = c

	// Add to the loop's poll to start polling for events
	l.poll.AddReadWrite(c.fd)

	return nil
}

type server struct {
	events   Events             // user events
	loops    []*loop            // all the loops
	lns      []*listener        // all the listeners
	wg       sync.WaitGroup     // loop close waitgroup
	cond     *sync.Cond         // shutdown signaler
	balance  LoadBalance        // load balancing method
	accepted uintptr            // accept counter
	dialed   uintptr            // dialed counter
	tch      chan time.Duration // ticker channel

	//ticktm   time.Time      // next tick time
}

type loop struct {
	idx     int            // loop index in the server loops list
	poll    *internal.Poll // epoll or kqueue
	packet  []byte         // read packet buffer
	fdconns map[int]*conn  // loop connections fd -> conn
	count   int32          // connection count

	Helpers LoopHelpers
}

func (s *server) SignalShutdown() error {
	s.signalShutdown()
	return nil
}

func (s *server) WaitForShutdown() error {
	s.waitForShutdown()
	return nil
}

// waitForShutdown waits for a signal to shutdown
func (s *server) waitForShutdown() {
	s.cond.L.Lock()
	s.cond.Wait()
	s.cond.L.Unlock()
}

// signalShutdown signals a shutdown an begins server closing
func (s *server) signalShutdown() {
	s.cond.L.Lock()
	s.cond.Signal()
	s.cond.L.Unlock()
}

func serve(events Events, listeners []*listener) error {
	// figure out the correct number of loops/goroutines to use.
	numLoops := events.NumLoops
	if numLoops <= 0 {
		if numLoops == 0 {
			numLoops = 1
		} else {
			numLoops = runtime.NumCPU()
		}
	}

	s := &server{}
	s.events = events
	s.lns = listeners
	s.cond = sync.NewCond(&sync.Mutex{})
	s.balance = events.LoadBalance
	s.tch = make(chan time.Duration)

	//println("-- server starting")
	if s.events.Serving != nil {
		var svr Server
		svr.NumLoops = numLoops
		svr.Addrs = make([]net.Addr, len(listeners))
		for i, ln := range listeners {
			svr.Addrs[i] = ln.lnaddr
		}
		svr.Dial = s.Dial
		svr.DialSockAddr = s.DialSockAddr
		svr.Shutdown = s.SignalShutdown
		svr.WaitForShutdown = s.WaitForShutdown
		action := s.events.Serving(svr)
		switch action {
		case None:
		case Shutdown:
			return nil
		}
	}

	defer func() {
		// wait on a signal for shutdown
		s.waitForShutdown()

		// notify all loops to close by closing all listeners
		for _, l := range s.loops {
			l.poll.Trigger(errClosing)
		}

		// wait on all loops to complete reading events
		s.wg.Wait()

		// close loops and all outstanding connections
		for _, l := range s.loops {
			for _, c := range l.fdconns {
				loopCloseConn(s, l, c, nil)
			}
			l.poll.Close()
		}
		//println("-- server stopped")
	}()

	// create loops locally and bind the listeners.
	for i := 0; i < numLoops; i++ {
		l := &loop{
			idx:     i,
			poll:    internal.OpenPoll(),
			packet:  make([]byte, 0xFFFF),
			fdconns: make(map[int]*conn),
		}
		for _, ln := range listeners {
			l.poll.AddRead(ln.fd)
		}
		s.loops = append(s.loops, l)
	}
	// start loops in background
	s.wg.Add(len(s.loops))
	for _, l := range s.loops {
		go loopRun(s, l)
	}
	return nil
}

func loopCloseConn(s *server, l *loop, c *conn, err error) error {
	atomic.AddInt32(&l.count, -1)
	delete(l.fdconns, c.fd)
	syscall.Close(c.fd)
	if s.events.Closed != nil {
		switch s.events.Closed(c, err) {
		case None:
		case Shutdown:
			return errClosing
		}
	}
	return nil
}

func loopDetachConn(s *server, l *loop, c *conn, err error) error {
	if s.events.Detached == nil {
		return loopCloseConn(s, l, c, err)
	}
	l.poll.ModDetach(c.fd)

	atomic.AddInt32(&l.count, -1)
	delete(l.fdconns, c.fd)
	if err := syscall.SetNonblock(c.fd, false); err != nil {
		return err
	}
	switch s.events.Detached(c, &detachedConn{fd: c.fd}) {
	case None:
	case Shutdown:
		return errClosing
	}
	return nil
}

func loopNote(s *server, l *loop, note interface{}) error {
	var err error
	switch v := note.(type) {
	case time.Duration:
		delay, action := s.events.Tick(l.idx)
		switch action {
		case None:
		case Shutdown:
			err = errClosing
		}
		s.tch <- delay
	case error: // shutdown
		err = v
	}
	return err
}

func loopRun(s *server, l *loop) {
	defer func() {
		//fmt.Println("-- loop stopped --", l.idx)
		s.signalShutdown()
		s.wg.Done()
	}()

	if l.idx == 0 && s.events.Tick != nil {
		go loopTicker(s, l)
	}

loop:
//fmt.Println("-- loop started --", l.idx)
	for {
		if err := l.poll.Wait(func(fd int, note interface{}) error {
			if fd == 0 {
				return loopNote(s, l, note)
			}
			c := l.fdconns[fd]
			switch {
			case c == nil:
				return loopAccept(s, l, fd)
			case !c.opened:
				return loopOpened(s, l, c)
			case len(c.out) > 0:
				return loopWrite(s, l, c)
			case c.action != None:
				return loopAction(s, l, c)
			default:
				return loopRead(s, l, c)
			}
		}); err != nil {
			break loop
		} else {
			break loop
		}
	}

}

func loopTicker(s *server, l *loop) {
	for {
		if err := l.poll.Trigger(time.Duration(0)); err != nil {
			break
		}
		time.Sleep(<-s.tch)
	}
}

func loopAccept(s *server, l *loop, fd int) error {
	for i, ln := range s.lns {
		if ln.fd == fd {
			if len(s.loops) > 1 {
				switch s.balance {
				case LeastConnections:
					n := atomic.LoadInt32(&l.count)
					for _, lp := range s.loops {
						if lp.idx != l.idx {
							if atomic.LoadInt32(&lp.count) < n {
								return nil // do not accept
							}
						}
					}
				case RoundRobin:
					if int(atomic.LoadUintptr(&s.accepted))%len(s.loops) != l.idx {
						return nil // do not accept
					}
					atomic.AddUintptr(&s.accepted, 1)
				}
			}
			if ln.pconn != nil {
				return loopUDPRead(s, l, i, fd)
			}
			nfd, sa, err := syscall.Accept(fd)
			if err != nil {
				if err == syscall.EAGAIN {
					return nil
				}
				return err
			}
			if err := syscall.SetNonblock(nfd, true); err != nil {
				return err
			}
			c := &conn{fd: nfd, sa: sa, lnidx: i, loop: l}
			l.fdconns[c.fd] = c
			l.poll.AddReadWrite(c.fd)
			atomic.AddInt32(&l.count, 1)
			break
		}
	}
	return nil
}

func loopUDPRead(s *server, l *loop, lnidx, fd int) error {
	n, sa, err := syscall.Recvfrom(fd, l.packet, 0)
	if err != nil || n == 0 {
		return nil
	}
	if s.events.Data != nil {
		var sa6 syscall.SockaddrInet6
		switch sa := sa.(type) {
		case *syscall.SockaddrInet4:
			sa6.ZoneId = 0
			sa6.Port = sa.Port
			for i := 0; i < 12; i++ {
				sa6.Addr[i] = 0
			}
			sa6.Addr[12] = sa.Addr[0]
			sa6.Addr[13] = sa.Addr[1]
			sa6.Addr[14] = sa.Addr[2]
			sa6.Addr[15] = sa.Addr[3]
		case *syscall.SockaddrInet6:
			sa6 = *sa
		}
		c := &conn{loop: l}
		c.addrIndex = lnidx
		c.localAddr = s.lns[lnidx].lnaddr
		c.remoteAddr = internal.SockaddrToAddr(&sa6)
		in := append([]byte{}, l.packet[:n]...)
		out, action := s.events.Data(c, in)
		if len(out) > 0 {
			if s.events.PreWrite != nil {
				s.events.PreWrite()
			}
			syscall.Sendto(fd, out, 0, sa)
		}
		switch action {
		case Shutdown:
			return errClosing
		}
	}
	return nil
}

func loopOpened(s *server, l *loop, c *conn) error {
	c.opened = true
	c.addrIndex = c.lnidx
	c.localAddr = s.lns[c.lnidx].lnaddr
	c.remoteAddr = internal.SockaddrToAddr(c.sa)
	if s.events.Opened != nil {
		out, opts, action := s.events.Opened(c)
		if len(out) > 0 {
			c.out = append([]byte{}, out...)
		}
		c.action = action
		c.reuse = opts.ReuseInputBuffer
		if opts.TCPKeepAlive > 0 {
			if _, ok := s.lns[c.lnidx].ln.(*net.TCPListener); ok {
				internal.SetKeepAlive(c.fd, int(opts.TCPKeepAlive/time.Second))
			}
		}
	}
	if len(c.out) == 0 && c.action == None {
		l.poll.ModRead(c.fd)
	}
	return nil
}

func loopWrite(s *server, l *loop, c *conn) error {
	if s.events.PreWrite != nil {
		s.events.PreWrite()
	}
	n, err := syscall.Write(c.fd, c.out)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return loopCloseConn(s, l, c, err)
	}
	if n == len(c.out) {
		c.out = nil
	} else {
		c.out = c.out[n:]
	}
	if len(c.out) == 0 && c.action == None {
		l.poll.ModRead(c.fd)
	}
	return nil
}

func loopAction(s *server, l *loop, c *conn) error {
	switch c.action {
	default:
		c.action = None
	case Close:
		return loopCloseConn(s, l, c, nil)
	case Shutdown:
		return errClosing
	case Detach:
		return loopDetachConn(s, l, c, nil)
	}
	if len(c.out) == 0 && c.action == None {
		l.poll.ModRead(c.fd)
	}
	return nil
}

func loopRead(s *server, l *loop, c *conn) error {
	var in []byte
	n, err := syscall.Read(c.fd, l.packet)
	if n == 0 || err != nil {
		if err == syscall.EAGAIN {
			if s.events.Data != nil {
				out, action := s.events.Data(c, nil)
				c.action = action
				if len(out) > 0 {
					c.out = append([]byte{}, out...)
					l.poll.ModReadWrite(c.fd)
				}
			}
			return nil
		}
		return loopCloseConn(s, l, c, err)
	}
	in = l.packet[:n]
	if !c.reuse {
		in = append([]byte{}, in...)
	}
	if s.events.Data != nil {
		out, action := s.events.Data(c, in)
		c.action = action
		if len(out) > 0 {
			c.out = append([]byte{}, out...)
		}
	}
	if len(c.out) != 0 || c.action != None {
		l.poll.ModReadWrite(c.fd)
	}
	return nil
}

type detachedConn struct {
	fd int
}

func (c *detachedConn) Close() error {
	err := syscall.Close(c.fd)
	if err != nil {
		return err
	}
	c.fd = -1
	return nil
}

func (c *detachedConn) Read(p []byte) (n int, err error) {
	return syscall.Read(c.fd, p)
}

func (c *detachedConn) Write(p []byte) (n int, err error) {
	n = len(p)
	for len(p) > 0 {
		nn, err := syscall.Write(c.fd, p)
		if err != nil {
			return n, err
		}
		p = p[nn:]
	}
	return n, nil
}

func (ln *listener) close() {
	if ln.fd != 0 {
		syscall.Close(ln.fd)
	}
	if ln.f != nil {
		ln.f.Close()
	}
	if ln.ln != nil {
		ln.ln.Close()
	}
	if ln.pconn != nil {
		ln.pconn.Close()
	}
	if ln.network == "unix" {
		os.RemoveAll(ln.addr)
	}
}

// system takes the net listener and detaches it from it's parent
// event loop, grabs the file descriptor, and makes it non-blocking.
func (ln *listener) system() error {
	var err error
	switch netln := ln.ln.(type) {
	case nil:
		switch pconn := ln.pconn.(type) {
		case *net.UDPConn:
			ln.f, err = pconn.File()
		}
	case *net.TCPListener:
		ln.f, err = netln.File()
	case *net.UnixListener:
		ln.f, err = netln.File()
	}
	if err != nil {
		ln.close()
		return err
	}
	ln.fd = int(ln.f.Fd())
	return syscall.SetNonblock(ln.fd, true)
}

func reuseportListenPacket(proto, addr string) (l net.PacketConn, err error) {
	return reuseport.ListenPacket(proto, addr)
}

func reuseportListen(proto, addr string) (l net.Listener, err error) {
	return reuseport.Listen(proto, addr)
}
