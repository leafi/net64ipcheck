package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

func prepend(xs []byte) []byte {
	var xb1 = append([]byte{}, 0x43, 0x4f, 0x4e, 0x43)
	return append(xb1, xs...)
}

func main() {
	fmt.Println("#######\nStarting net64ipcheck.\n#######")

	// port 6460

	listenAddr, err := net.ResolveUDPAddr("udp", ":6460")
	if err != nil {
		panic("Failed to construct empty/any udp address for listening on port 6460")
	}

	ln, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		panic("FAILED to listen on udp port 6460 - is this process already running?")
	}
	defer ln.Close()

	buf := make([]byte, 1024)

	fmt.Println("Starting listen loop.")

	for {
		n, raddr, err := ln.ReadFromUDP(buf)

		if err != nil {
			fmt.Println("Read error: ", err)
		} else if n != 400 {
			// packet must be 400 bytes; larger than necessary to prevent attacks where Outgoing >>> Incoming
			// (i.e. using us to spam the ostensible sender with more data)
			fmt.Println("Bad fmt packet from ", raddr, " (not 400 bytes)")
		} else if string(buf[0:12]) != "net64ipc0000" {
			fmt.Println("Bad fmt packet from ", raddr, " (expected net64ipc0000)")
			go func(raddr *net.UDPAddr) {
				ln.WriteToUDP(prepend([]byte("BADVER")), raddr)
			}(raddr)
		} else {
			// parse server port from packet
			srvport := binary.LittleEndian.Uint16(buf[12:14])

			// only allow srvport == raddr.Port (Net64 tool does this, and it makes things simpler.)
			if int(srvport) == raddr.Port {
				go func(raddr *net.UDPAddr, srvport uint16) {
					_, err := ln.WriteToUDP(prepend([]byte("OK "+raddr.IP.String())), raddr)

					if err != nil {
						fmt.Println(err, "\n^ Failed to write OK for given port ", srvport, " (raddr ", raddr, ")")
					}
				}(raddr, srvport)
			} else {
				fmt.Println("srvport != raddr.Port; ignoring message!")
			}
		}
	}

}
