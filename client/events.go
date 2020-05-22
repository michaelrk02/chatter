package main

import (
    "context"
    "fmt"
    "github.com/golang/protobuf/ptypes/wrappers"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "io"
    "sync/atomic"
    "time"
)

func pollClientConnectEvents() {
    var err error
    stream, err := chatterEventsCln.ListenClientConnect(context.Background(), &wrappers.UInt32Value{Value: clientId})
    if err != nil {
        panic(err)
    }
    for atomic.LoadInt32(&running) == 1 {
        evt, err := stream.Recv()
        if err != nil {
            if err == io.EOF || status.Code(err) == codes.NotFound {
                break
            } else {
                panic(err)
            }
        }
        outMu.Lock()
        fmt.Printf("[%s] %s has entered the chat\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Nickname)
        outMu.Unlock()
    }
}

func pollClientDisconnectEvents() {
    var err error
    stream, err := chatterEventsCln.ListenClientDisconnect(context.Background(), &wrappers.UInt32Value{Value: clientId})
    if err != nil {
        panic(err)
    }
    for atomic.LoadInt32(&running) == 1 {
        evt, err := stream.Recv()
        if err != nil {
            if err == io.EOF || status.Code(err) == codes.NotFound {
                break
            } else {
                panic(err)
            }
        }
        outMu.Lock()
        fmt.Printf("[%s] %s has left the chat\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Nickname)
        outMu.Unlock()
    }
}

func pollClientMessageEvents() {
    var err error
    stream, err := chatterEventsCln.ListenClientMessage(context.Background(), &wrappers.UInt32Value{Value: clientId})
    if err != nil {
        panic(err)
    }
    for atomic.LoadInt32(&running) == 1 {
        evt, err := stream.Recv()
        if err != nil {
            if err == io.EOF || status.Code(err) == codes.NotFound {
                break
            } else {
                panic(err)
            }
        }
        outMu.Lock()
        fmt.Printf("[%s] %s: %s\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Nickname, evt.Contents)
        outMu.Unlock()
    }
}

func pollServerMessageEvents() {
    var err error
    stream, err := chatterEventsCln.ListenServerMessage(context.Background(), &wrappers.UInt32Value{Value: clientId})
    if err != nil {
        panic(err)
    }
    for atomic.LoadInt32(&running) == 1 {
        evt, err := stream.Recv()
        if err != nil {
            if err == io.EOF || status.Code(err) == codes.NotFound {
                break
            } else {
                panic(err)
            }
        }
        outMu.Lock()
        fmt.Printf("[%s] SERVER: %s\n", time.Unix(evt.Timestamp, 0).Local().Format(time.Kitchen), evt.Contents)
        outMu.Unlock()
    }
}

