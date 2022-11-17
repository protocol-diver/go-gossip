package gogossip

import (
	"net"
)

type Transport interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
}

func multicastWithRawAddress(transport Transport, buf []byte, peers []string) {
	addrs := make([]*net.UDPAddr, 0, len(peers))
	for _, peer := range peers {
		addr, err := net.ResolveUDPAddr("udp", peer)
		if err != nil {
			// TODO: returns error or have tolerance?
			continue
		}
		addrs = append(addrs, addr)
	}
	multicastWithAddress(transport, buf, addrs)
}

func multicastWithAddress(transport Transport, buf []byte, addrs []*net.UDPAddr) {
	for _, addr := range addrs {
		if _, err := transport.WriteToUDP(buf, addr); err != nil {
			continue
		}
	}
}
