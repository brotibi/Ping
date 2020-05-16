package main

/*
	My implementation of a ICMP Ping program in Golang


	Root Priveleges are required to run this application so to run this type in "sudo ./main <hostname>""
	or is you want to set TTL "sudo ./main -ttl <Seconds to live> <hostname>""
*/

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ICMPIPV4 = 1  // ICMP Protocol for IPV6
	ICMPIPV6 = 58 // Ended up not implementing IPV6
)

var hostname = "8.8.8.8"
var totalPackets = 0
var packetsRead = 0
var listenAddr = "0.0.0.0"
var ttlVar int

func getArgs() { // Gets the arguments of the command line

	flag.IntVar(
		&ttlVar, "ttl", -1,
		"This is if you want to enable TTL",
	)
	flag.Parse()

	//os.Args[1]
	hostname = flag.Arg(0)

	fmt.Println(ttlVar)
}

func Ping() { // The main pinging function

	totalPackets++
	conn, err := icmp.ListenPacket("ip4:icmp", listenAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	dest, err := net.ResolveIPAddr("ip4", hostname)

	if err != nil {
		fmt.Println(err)
		return
	}

	mesg := icmp.Message{ // Creates the ICMP Echo Message
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte(""),
		},
	}
	if ttlVar != -1 {
		conn.IPv4PacketConn().SetTTL(ttlVar) // Sets the TTL
	}
	b, err := mesg.Marshal(nil)

	if err != nil {
		fmt.Println(dest, err)
	}

	start := time.Now() // We then record the time right before we send it
	// Send it
	size, err := conn.WriteTo(b, dest)

	if err != nil {
		fmt.Println(dest, err)
	} else if size != len(b) {
		fmt.Println("got ", size, " want", "len(b)")
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	//if ttlVar == -1 {
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	if err != nil {
		fmt.Println(dest, err)
	}
	size, peer, err := conn.ReadFrom(reply)
	if err != nil && ttlVar != -1 {
		fmt.Println(dest, err, "Time To Live", ttlVar)
	}
	duration := time.Since(start) // We then determine the amount the time that has passed

	// We then parse the message
	retMsg, err := icmp.ParseMessage(ICMPIPV4, reply[:size])
	if err != nil {
		fmt.Println(dest, err)
		return
	}
	switch retMsg.Type {
	case ipv4.ICMPTypeEchoReply:
		packetsRead++
		//fmt.Println(packetsRead, totalPackets)
		read := float64(packetsRead)
		total := float64(totalPackets)
		fmt.Println("Pinging", hostname, duration, "Packet Loss: ", 100-(read/total)*100, "%  ", totalPackets-packetsRead, "packets lost total")
	default:
		fmt.Println("got ", retMsg, " from ", peer, ". wanted a echo reply")
	}
}

func main() {
	getArgs()
	for true { // The infinite loop
		Ping()
	}

}
