package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

///////////////////////////////////////////////////////////////////////////////////////////////////

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// TODO: update this to use crypto/rand.Text(), it's in Go 1.24 but gccgo on the Raspberry Pi
// doesn't have the latest updates yet.
func randomString() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 26)
	for i, _ := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func getLocalIP() (string, error) {
	fmt.Println("checking local IP addresses...")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("could not get host's addresses")
	}

	for i := 0; i < len(addrs); i++ {
		localAddr := addrs[i].String()
		if strings.HasPrefix(localAddr, "192.168.1.") {
			localAddr = strings.Split(localAddr, "/")[0]

			// make sure this address is usable (connected to the internet)
			if ping(localAddr, "www.google.com:443") {
				return localAddr, nil
			}
		}
	}

	return "", errors.New("could not find local address")
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func ping(localAddress, remoteAddress string) bool {
	localAddress += ":0"
	// not an actual ping (ICMP echo), but good enough for now
	fmt.Println("'pinging' using local address", localAddress)

	lAddress, err := net.ResolveTCPAddr("tcp", localAddress)
	if err != nil {
		return false
	}

	rAddress, err := net.ResolveTCPAddr("tcp", remoteAddress)
	if err != nil {
		return false
	}

	conn, err := net.DialTCP("tcp", lAddress, rAddress)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println(conn.LocalAddr().String(), "is connected to the internet")
	conn.Close()

	return true
}
