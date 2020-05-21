package main

import (
    "bufio"
    "context"
    "fmt"
    "github.com/golang/protobuf/ptypes/empty"
    "github.com/golang/protobuf/ptypes/wrappers"
    "github.com/michaelrk02/chatter"
    "google.golang.org/grpc"
    "io"
    "os"
    "sync/atomic"
    "time"
)

func main() {
    var err error

    var address string
    fmt.Printf("Enter server address: ")
    fmt.Scanf("%s", &address)

    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    chatterCln = chatter.NewChatterClient(conn)
    chatterEventsCln = chatter.NewChatterEventsClient(conn)

    var nickname string
    fmt.Printf("Enter your nickname: ")
    fmt.Scanf("%s", &nickname)

    clientIdValue, err := chatterCln.Connect(context.Background(), &wrappers.StringValue{Value: nickname})
    if err != nil {
        panic(err)
    }
    clientId = clientIdValue.Value
    running = 1
    fmt.Println("Type `list` to list clients, `send` to send a message, or `quit` to disconnect")
    go pollClientConnectEvents()
    go pollClientDisconnectEvents()
    go pollClientMessageEvents()
    go pollServerMessageEvents()
    for atomic.LoadInt32(&running) == 1 {
        var cmd string
        fmt.Printf("$ ")
        fmt.Scanf("%s", &cmd)

        if cmd == "list" {
            stream, err := chatterCln.GetClients(context.Background(), &empty.Empty{})
            if err != nil {
                outMu.Lock()
                fmt.Printf("Unable to list clients: %s\n", err)
                outMu.Unlock()
                continue
            }
            outMu.Lock()
            fmt.Println("=== List of clients ===")
            outMu.Unlock()
            for {
                cl, err := stream.Recv()
                if err != nil {
                    if err == io.EOF {
                        break
                    } else {
                        outMu.Lock()
                        fmt.Printf("List clients error: %s\n", err)
                        outMu.Unlock()
                        break
                    }
                }
                outMu.Lock()
                fmt.Printf("- %s [joined on %s]\n", cl.Nickname, time.Unix(cl.JoinTimestamp, 0).Local().Format(time.RFC3339))
                outMu.Unlock()
            }
        }
        if cmd == "send" {
            rd := bufio.NewReader(os.Stdin)

            outMu.Lock()
            fmt.Printf("Message: ")
            outMu.Unlock()
            lineBuf, _, _ := rd.ReadLine()
            line := string(lineBuf)

            _, err = chatterCln.Message(context.Background(), &chatter.MessageDescriptor{ClientId: clientId, Contents: line})
            if err != nil {
                outMu.Lock()
                fmt.Printf("Unable to send message: %s\n", err)
                outMu.Unlock()
                continue
            }
        }
        if cmd == "quit" {
            _, err = chatterCln.Disconnect(context.Background(), &wrappers.UInt32Value{Value: clientId})
            if err != nil {
                outMu.Lock()
                fmt.Printf("Unable to disconnect: %s\n", err)
                outMu.Unlock()
                continue
            }
            atomic.StoreInt32(&running, 0)
        }
    }
}

