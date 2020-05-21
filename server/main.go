package main

import (
    "bufio"
    "fmt"
    "github.com/michaelrk02/chatter"
    "google.golang.org/grpc"
    "math/rand"
    "net"
    "os"
    "time"
)

func main() {
    rand.Seed(time.Now().Unix())

    var port int
    fmt.Printf("Enter the host port: ")
    fmt.Scanf("%d", &port)

    var err error

    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        panic(err)
    }
    defer lis.Close()

    chatterSrv = new(chatterServer).Init()
    chatterEventsSrv = new(chatterEventsServer).Init()

    srv := grpc.NewServer()
    chatter.RegisterChatterServer(srv, chatterSrv)
    chatter.RegisterChatterEventsServer(srv, chatterEventsSrv)

    fmt.Println("Type `send` to send a message, `list` to list clients, or `quit` to quit")
    go (func() {
        running := true
        for running {
            var cmd string
            outMu.Lock()
            fmt.Printf("$ ")
            outMu.Unlock()
            fmt.Scanf("%s", &cmd)

            if cmd == "send" {
                rd := bufio.NewReader(os.Stdin)

                outMu.Lock()
                fmt.Printf("Message: ")
                outMu.Unlock()
                lineBuf, _, _ := rd.ReadLine()
                line := string(lineBuf)

                chatterEventsSrv.ServerMessage.Push(&chatterEventsSrv.Mu, &chatter.ServerMessageEvent{Contents: line, Timestamp: time.Now().UTC().Unix()})

                outMu.Lock()
                fmt.Printf("SERVER: %s\n", line)
                outMu.Unlock()
            }
            if cmd == "list" {
                chatterSrv.ListClients()
            }
            if cmd == "quit" {
                running = false
            }
        }
        srv.Stop()
    })()

    err = srv.Serve(lis)
    if err != nil {
        panic(err)
    }
}

