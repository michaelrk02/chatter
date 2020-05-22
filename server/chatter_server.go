package main

import (
    "context"
    "fmt"
    "github.com/golang/protobuf/ptypes/empty"
    "github.com/golang/protobuf/ptypes/wrappers"
    "github.com/michaelrk02/chatter"
    "io"
    "math/rand"
    "sync"
    "time"
)

type chatterServer struct {
    clientsMu sync.Mutex
    clients map[uint32]*client
}

func (s *chatterServer) Init() *chatterServer {
    s.clients = make(map[uint32]*client)

    return s
}

func (s *chatterServer) ListClients() {
    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()

    outMu.Lock()
    defer outMu.Unlock()

    fmt.Println("=== List of clients ===")
    for id, cl := range s.clients {
        fmt.Printf("%s (ID: %d)\n", cl.Nickname, id)
    }
}

func (s *chatterServer) checkExistency(clientId uint32) error {
    if _, idExists := s.clients[clientId]; !idExists {
        return chatter.CLIENT_NOT_FOUND
    }
    return nil
}

func (s *chatterServer) GetClients(req *empty.Empty, srv chatter.Chatter_GetClientsServer) error {
    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()
    for _, client := range s.clients {
        srv.Send(&chatter.ClientInfo{Nickname: client.Nickname, JoinTimestamp: client.JoinTimestamp})
    }
    return nil
}

func (s *chatterServer) Connect(ctx context.Context, req *wrappers.StringValue) (*wrappers.UInt32Value, error) {
    if req.Value == "SERVER" {
        return nil, chatter.INVALID_NICKNAME
    }

    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()

    nicknameUsed := false
    for _, client := range s.clients {
        if req.Value == client.Nickname {
            nicknameUsed = true
            break
        }
    }
    if nicknameUsed {
        return nil, chatter.NICKNAME_USED
    }

    var id uint32
    for {
        id = rand.Uint32()
        if s.checkExistency(id) != nil {
            break
        }
    }
    cl := new(client).Init(id, req.Value)
    s.clients[id] = cl
    chatterEventsSrv.RegisterClient(id)

    outMu.Lock()
    defer outMu.Unlock()
    fmt.Printf("%s (ID: %d) has entered the chat\n", req.Value, id)

    chatterEventsSrv.ClientConnect.Push(&chatterEventsSrv.Mu, &chatter.ClientConnectionEvent{Nickname: req.Value, Timestamp: cl.JoinTimestamp})

    return &wrappers.UInt32Value{Value: id}, nil
}

func (s *chatterServer) Disconnect(ctx context.Context, req *wrappers.UInt32Value) (*empty.Empty, error) {
    var err error

    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()

    if err = s.checkExistency(req.Value); err != nil {
        return nil, err
    }
    cl := s.clients[req.Value]

    chatterEventsSrv.RevokeClient(req.Value)
    delete(s.clients, req.Value)

    outMu.Lock()
    defer outMu.Unlock()
    fmt.Printf("%s (ID: %d) has left the chat\n", cl.Nickname, req.Value)

    chatterEventsSrv.ClientDisconnect.Push(&chatterEventsSrv.Mu, &chatter.ClientConnectionEvent{Nickname: cl.Nickname, Timestamp: time.Now().UTC().Unix()})

    return &empty.Empty{}, nil
}

func (s *chatterServer) Message(ctx context.Context, req *chatter.MessageDescriptor) (*empty.Empty, error) {
    var err error

    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()

    if err = s.checkExistency(req.ClientId); err != nil {
        return nil, err
    }
    cl := s.clients[req.ClientId]

    outMu.Lock()
    defer outMu.Unlock()
    fmt.Printf("%s (ID: %d): %s\n", cl.Nickname, req.ClientId, req.Contents)

    chatterEventsSrv.ClientMessage.Push(&chatterEventsSrv.Mu, &chatter.ClientMessageEvent{Nickname: cl.Nickname, Contents: req.Contents, Timestamp: time.Now().UTC().Unix()})

    return &empty.Empty{}, nil
}

func (s *chatterServer) Maintain(srv chatter.Chatter_MaintainServer) error {
    var err error

    clientIdValue, err := srv.Recv()
    if err != nil {
        return err
    }
    s.clientsMu.Lock()
    if err = s.checkExistency(clientIdValue.Value); err != nil {
        s.clientsMu.Unlock()
        return err
    }
    cl := s.clients[clientIdValue.Value]
    s.clientsMu.Unlock()

    _, err = srv.Recv()
    if err != nil && err != io.EOF {
        chatterEventsSrv.RevokeClient(clientIdValue.Value)
        s.clientsMu.Lock()
        delete(s.clients, clientIdValue.Value)
        s.clientsMu.Unlock()

        outMu.Lock()
        defer outMu.Unlock()
        fmt.Printf("%s (ID: %d) disconnected\n", cl.Nickname, clientIdValue.Value)

        chatterEventsSrv.ServerMessage.Push(&chatterEventsSrv.Mu, &chatter.ServerMessageEvent{Contents: fmt.Sprintf("%s disconnected\n", cl.Nickname), Timestamp: time.Now().UTC().Unix()})

        return err
    }

    return nil
}

