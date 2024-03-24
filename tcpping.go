package main

import (
	"bufio"
	"encoding/binary"
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
	log.SetFlags(log.Lmicroseconds)
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
	sbuf := make([]byte, *sz)
	rbuf := make([]byte, *sz)
	for range tick.C {
		binary.LittleEndian.PutUint64(sbuf, uint64(*sz))
		t, _ := time.Now().MarshalBinary()
		copy(sbuf[8:], t)
		n, err := conn.Write(sbuf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("wrote", n, "bytes")
		n, err = io.ReadFull(conn, rbuf)
		if err != nil {
			log.Fatal(err)
		}
		var start time.Time
		const timeStampLen = 15
		if len(rbuf) < 8+timeStampLen {
			log.Fatalln("rbuf too short")
		}
		err = start.UnmarshalBinary(rbuf[8 : 8+timeStampLen])
		if err != nil {
			log.Fatalln("could not parse time", err)
		}
		log.Println("read", n, "bytes, delay", time.Since(start))
	}

}

func doServer() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening on", l.Addr())
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

type lreader struct {
	r io.Reader
}

func (l *lreader) Read(buf []byte) (int, error) {
	n, err := l.r.Read(buf)
	log.Println("read", n, "bytes from socket")
	if err != nil {
		log.Println(err)
	}
	return n, err
}

func handleConnection(c net.Conn) {
	log.Println("got connection from", c.RemoteAddr())
	bc := bufio.NewReaderSize(&lreader{c}, 64*1028)
	szBuf := make([]byte, 8)

	buf := make([]byte, *sz)
	for {
		_, err := io.ReadFull(bc, szBuf)
		if err != nil {
			log.Println("reading size", err)
		}
		if err == io.EOF {
			return
		}
		sz := binary.LittleEndian.Uint64(szBuf)
		log.Println("reading", sz)
		if len(buf) != int(sz) {
			buf = make([]byte, int(sz))
		}
		n, err := io.ReadFull(bc, buf[8:])
		if err != nil {
			log.Fatal(err)
		}
		log.Println("got", n+8, "bytes")
		copy(buf, szBuf)
		c.Write(buf)
	}
}
