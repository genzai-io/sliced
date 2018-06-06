package evio

import (
	"syscall"
	"net"
)


// Resolve resolves an evio address and retuns a sockaddr for socket
// connection to external servers.
func Resolve(addr string) (sa syscall.Sockaddr, err error) {
	network, address, _, _ := parseAddr(addr)
	var taddr net.Addr
	switch network {
	default:
		return nil, net.UnknownNetworkError(network)
	case "unix":
		taddr = &net.UnixAddr{Net: "unix", Name: address}
	case "tcp", "tcp4", "tcp6":
		// use the stdlib resolver because it's good.
		taddr, err = net.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
	}
	switch taddr := taddr.(type) {
	case *net.UnixAddr:
		sa = &syscall.SockaddrUnix{Name: taddr.Name}
	case *net.TCPAddr:
		switch len(taddr.IP) {
		case 0:
			var sa4 syscall.SockaddrInet4
			sa4.Port = taddr.Port
			sa = &sa4
		case 4:
			var sa4 syscall.SockaddrInet4
			copy(sa4.Addr[:], taddr.IP[:])
			sa4.Port = taddr.Port
			sa = &sa4
		case 16:
			var sa6 syscall.SockaddrInet6
			copy(sa6.Addr[:], taddr.IP[:])
			sa6.Port = taddr.Port
			sa = &sa6
		}
	}
	return sa, nil
}

func SockaddrToAddr(sa syscall.Sockaddr) net.Addr {
	var a net.Addr
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
		}
	case *syscall.SockaddrInet6:
		var zone string
		if sa.ZoneId != 0 {
			if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
				zone = ifi.Name
			}
		}
		if zone == "" && sa.ZoneId != 0 {
		}
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
			Zone: zone,
		}
	case *syscall.SockaddrUnix:
		a = &net.UnixAddr{Net: "unix", Name: sa.Name}
	}
	return a
}
