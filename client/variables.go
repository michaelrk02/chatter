package main

import (
    "github.com/michaelrk02/chatter"
    "sync"
)

var chatterCln chatter.ChatterClient
var chatterEventsCln chatter.ChatterEventsClient

var clientId uint32
var running int32
var disconnecting int32 = -1

var outMu sync.Mutex

