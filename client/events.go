package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/golang/protobuf/ptypes/wrappers"
    "github.com/michaelrk02/chatter"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "io"
    "sync/atomic"
    "time"
)

type eventCallback func(evt interface{})

func pollEvents(event string, callback eventCallback) {
    var err error
    var stream interface{}
    var data interface{}
    arg := &wrappers.UInt32Value{Value: clientId}
    switch event {
    case "ClientConnect":
        stream, err = chatterEventsCln.ListenClientConnect(context.Background(), arg)
        data = new(chatter.ClientConnectionEvent)
    case "ClientDisconnect":
        stream, err = chatterEventsCln.ListenClientDisconnect(context.Background(), arg)
        data = new(chatter.ClientConnectionEvent)
    case "ClientMessage":
        stream, err = chatterEventsCln.ListenClientMessage(context.Background(), arg)
        data = new(chatter.ClientMessageEvent)
    case "ServerMessage":
        stream, err = chatterEventsCln.ListenServerMessage(context.Background(), arg)
        data = new(chatter.ServerMessageEvent)
    }
    if err != nil {
        panic(err)
    }
    for atomic.LoadInt32(&running) == 1 {
        err = stream.(grpc.ClientStream).RecvMsg(data)
        if err != nil {
            if err == io.EOF || status.Code(err) == codes.NotFound {
                break
            } else if status.Code(err) == codes.Unavailable {
                if atomic.LoadInt32(&running) == 1 {
                    atomic.StoreInt32(&running, 0)
                    panic(errors.New("server shutdown"))
                }
            } else {
                panic(err)
            }
        }
        callback(data)
    }
}

func pollClientConnectEvents() {
    pollEvents("ClientConnect", func(data interface{}) {
        evt := data.(*chatter.ClientConnectionEvent)
        outMu.Lock()
        fmt.Printf("[%s] %s has entered the chat\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Nickname)
        outMu.Unlock()
    })
}

func pollClientDisconnectEvents() {
    pollEvents("ClientDisconnect", func(data interface{}) {
        evt := data.(*chatter.ClientConnectionEvent)
        outMu.Lock()
        fmt.Printf("[%s] %s has left the chat\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Nickname)
        outMu.Unlock()
    })
}

func pollClientMessageEvents() {
    pollEvents("ClientMessage", func(data interface{}) {
        evt := data.(*chatter.ClientMessageEvent)
        outMu.Lock()
        fmt.Printf("[%s] %s: %s\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Nickname, evt.Contents)
        outMu.Unlock()
    })
}

func pollServerMessageEvents() {
    pollEvents("ServerMessage", func(data interface{}) {
        evt := data.(*chatter.ServerMessageEvent)
        outMu.Lock()
        fmt.Printf("[%s] SERVER: %s\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Contents)
        outMu.Unlock()
    })
}

