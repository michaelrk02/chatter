package main

import (
    "sync"
)

var chatterSrv *chatterServer
var chatterEventsSrv *chatterEventsServer

var outMu sync.Mutex

