package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"
)

var addr = flag.String("addr", ":9981", "address of transporter server")
var shell = flag.String("sh", "/bin/bash", "shell")
var id = flag.Int("id", -1, "connection id")

func doConn(conn net.Conn, id int) {
	defer conn.Close()

	fmt.Fprintf(conn, "CONN %d\r\n", id)

	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil {
		log.Print(err)
		return
	}
	line = strings.Trim(line, "\r\n")
	ret := strings.Fields(line)
	if ret[0] != "200" {
		log.Print(ret[1])
		return
	}
	log.Print(ret)
	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)
	log.Print("Connection closed")
}

func doReg(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintf(conn, "REG\r\n")

	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil {
		log.Print(err)
		return
	}
	line = strings.Trim(line, "\r\n")
	ret := strings.Fields(line)
	if ret[0] != "200" {
		log.Print(ret[1])
		return
	}
	log.Printf("id: %s\n", ret[1])

	c := exec.Command(*shell)
	fd, err := pty.Start(c)

	if err != nil {
		log.Print(err)
		return
	}
	var w io.Writer
	w = io.MultiWriter(conn, os.Stdout)
	go func() {
		io.Copy(w, fd)
		conn.Close()
	}()
	io.Copy(fd, conn)
	// Write EOF
	fd.Write([]byte{4})
	c.Wait()
	log.Print("Connection closed")
}

func main() {
	flag.Parse()
	log.Print("Connecting ", *addr)
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Connected")

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatal(err)
	}

	defer terminal.Restore(0, oldState)

	if *id != -1 {
		doConn(conn, *id)
	} else {
		doReg(conn)
	}
}
