package main

import (
        "bufio"
        "flag"
        "fmt"
        "io"
        "log"
        "net"
        "strconv"
        "strings"
)

type peer struct {
        pipe io.ReadWriteCloser
}

var addr = flag.String("addr", ":9981", "address of listen server")
var m map[int]*peer
var n int

func handleConn(conn net.Conn) {
        buf := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
        line, err := buf.ReadString('\n')
        if err != nil {
                log.Print(err)
                return
        }
        line = strings.Trim(line, "\r\n")
        args := strings.Fields(line)

        if len(args) == 0 {
                fmt.Fprint(conn, "500 Arg error\r\n")
                return
        }

        var action string
        action = args[0]

        if action == "REG" {
                m[n] = &peer{pipe: conn}
                log.Printf("Register to %d", n)
                fmt.Fprintf(conn, "200 %d\r\n", n)
                n++
                return
        } else if action == "CONN" {
                if len(args) < 2 {
                        fmt.Fprint(conn, "500 missing id\r\n")
                        return
                }
                who, _ := strconv.Atoi(args[1])
                p, ok := m[who]
                if !ok {
                        fmt.Fprint(conn, "500 invalid id\r\n")
                        return
                }
                defer delete(m, who)
                fmt.Fprintf(conn, "200 OK\r\n")
                log.Printf("Connect to %d", who)
                go func() {
                        io.Copy(conn, p.pipe)
                        conn.Close()
                        p.pipe.Close()
                }()
                io.Copy(p.pipe, conn)
                p.pipe.Close()
                conn.Close()
                log.Print("Connection Closed")
        } else {
                fmt.Fprint(conn, "500 invalid command\r\n")
                conn.Close()
                log.Print("Connection Closed")
        }
}

func main() {
        flag.Parse()
        m = make(map[int]*peer)
        l, err := net.Listen("tcp", *addr)
        if err != nil {
                log.Fatal(err)
        }
        log.Printf("Listening on addr:%s", *addr)
        for {
                conn, err := l.Accept()
                if err != nil {
                        log.Print(err)
                        continue
                }
                log.Printf("Connection from addr:%s", conn.RemoteAddr())
                go handleConn(conn)
        }
}
