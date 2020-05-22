package main

import (
    "github.com/golang/protobuf/ptypes/wrappers"
    "github.com/michaelrk02/chatter"
    "google.golang.org/grpc"
    "sync"
)

type chatterEventsServer struct {
    Mu sync.Mutex

    ClientConnect eventQueueMap
    ClientDisconnect eventQueueMap
    ClientMessage eventQueueMap

    ServerMessage eventQueueMap

    mapping map[string]*eventQueueMap
}

func (s *chatterEventsServer) Init() *chatterEventsServer {
    s.ClientConnect = make(eventQueueMap)
    s.ClientDisconnect = make(eventQueueMap)
    s.ClientMessage = make(eventQueueMap)

    s.ServerMessage = make(eventQueueMap)

    s.mapping = make(map[string]*eventQueueMap)
    s.mapping["ClientConnect"] = &s.ClientConnect
    s.mapping["ClientDisconnect"] = &s.ClientDisconnect
    s.mapping["ClientMessage"] = &s.ClientMessage
    s.mapping["ServerMessage"] = &s.ServerMessage

    return s
}

func (s *chatterEventsServer) RegisterClient(clientId uint32) {
    s.Mu.Lock()
    defer s.Mu.Unlock()

    for _, queueMap := range s.mapping {
        (*queueMap)[clientId] = new(eventQueue).Init()
    }
}

func (s *chatterEventsServer) RevokeClient(clientId uint32) {
    s.Mu.Lock()
    defer s.Mu.Unlock()

    for _, queueMap := range s.mapping {
        delete(*queueMap, clientId)
    }
}

func (s *chatterEventsServer) listen(clientId uint32, event string, stream grpc.ServerStream) error {
    var err error
    queueMap := *s.mapping[event]
    for {
        s.Mu.Lock()
        if queue, queueOk := queueMap[clientId]; queueOk {
            evt := queue.Pop()
            if evt != nil {
                err = stream.SendMsg(evt)
            }
        } else {
            err = chatter.CLIENT_NOT_FOUND
        }
        s.Mu.Unlock()

        if err != nil {
            break
        }
    }
    return err    
}

func (s *chatterEventsServer) ListenClientConnect(req *wrappers.UInt32Value, srv chatter.ChatterEvents_ListenClientConnectServer) error {
    return s.listen(req.Value, "ClientConnect", srv)
}

func (s *chatterEventsServer) ListenClientDisconnect(req *wrappers.UInt32Value, srv chatter.ChatterEvents_ListenClientDisconnectServer) error {
    return s.listen(req.Value, "ClientDisconnect", srv)
}

func (s *chatterEventsServer) ListenClientMessage(req *wrappers.UInt32Value, srv chatter.ChatterEvents_ListenClientMessageServer) error {
    return s.listen(req.Value, "ClientMessage", srv)
}

func (s *chatterEventsServer) ListenServerMessage(req *wrappers.UInt32Value, srv chatter.ChatterEvents_ListenServerMessageServer) error {
    return s.listen(req.Value, "ServerMessage", srv)
}

