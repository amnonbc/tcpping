package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var (
	dest = flag.String("dest", "", "destination host")

	port   = flag.Int("port", 9876, "port")
	server = flag.Bool("server", false, "be server")
	delay  = flag.Duration("delay", time.Second, "delay")
	sz     = flag.Int("size", 10*1024, "size of payload")
)

func main() {
	flag.Parse()
	if *server {
		doServer()
	} else {
		doClient()
	}

}

func doClient() {
	var d net.Dialer

	conn, err := d.Dial("tcp", *dest)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	tick := time.NewTicker(*delay)
	buf := make([]byte, *sz)
	for range tick.C {
		start := time.Now()
		n, err := conn.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("wrote", n, "bytes")
		n, err = io.ReadFull(conn, buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("read", n, "bytes, delay", time.Since(start))
	}

}

func doServer() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	buf := make([]byte, *sz)
	for {
		n, err := io.ReadFull(c, buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("got", n, "bytes")
		c.Write(buf)
	}
}
