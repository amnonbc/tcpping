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
		n, err = conn.Read(buf)
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
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			io.Copy(c, c)
			// Shut down the connection.
			c.Close()
		}(conn)
	}
}

func handleConnection(c net.Conn) {
	buf := make([]byte, *sz)
	for {
		n, err := c.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("got", n, "bytes")
		c.Write(buf)
	}
}
