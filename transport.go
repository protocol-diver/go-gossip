package gogossip

import (
	"net"
	"net/netip"
)

type Transport interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)

	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	WriteMsgUDPAddrPort(b []byte, oob []byte, addr netip.AddrPort) (n int, oobn int, err error)
}
