package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

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
		} else if string(buf[0:11]) != "net64ipc0000" {
			fmt.Println("Bad fmt packet from ", raddr, " (expected net64ipc0000)")
			go func(raddr *net.UDPAddr) {
				c, err := net.DialUDP("udp", nil, raddr)
				defer c.Close()

				if err == nil {
					c.Write([]byte("BADVER"))
				}
			}(raddr)
		} else {
			// parse server port from packet
			srvport := binary.LittleEndian.Uint16(buf[12:13])
			go func(raddr *net.UDPAddr, srvport uint16) {
				srvAddr := &net.UDPAddr{
					IP:   raddr.IP,
					Port: int(srvport),
					Zone: raddr.Zone,
				}

				c, err := net.DialUDP("udp", nil, srvAddr)
				defer c.Close()

				if err != nil {
					fmt.Println("Failed to construct udp address for given port ", srvport, " (raddr ", raddr, " srvAddr ", srvAddr, ")")
				} else {
					c.Write([]byte("TEST"))

					c.SetReadDeadline(time.Now().Add(time.Second * 2))

					rbuf := make([]byte, 9)
					n, err := c.Read(rbuf)

					if err != nil {
						fmt.Println("  no/err read from sending TEST to ", srvAddr, " (raddr ", raddr, ")")
					} else if n != 8 {
						fmt.Println("  got response but a bad one from sending TEST to ", srvAddr, " (expected 8 bytes)")
					} else if string(rbuf[0:7]) != "TOAST000" {
						fmt.Println("  got response but bad one from sending TEST to ", srvAddr, " (expected TOAST000)")
					} else {
						c2, err := net.DialUDP("udp", nil, raddr)
						defer c2.Close()

						if err != nil {
							c2.Write([]byte("OK " + srvAddr.IP.String()))
						}
					}
				}
			}(raddr, srvport)
		}
	}

}
