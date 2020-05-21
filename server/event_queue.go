package main

import (
    "container/list"
    "sync"
)

type eventQueue struct {
    data *list.List
    mu sync.Mutex
}

func (q *eventQueue) Init() *eventQueue {
    q.data = list.New()

    return q
}

func (q *eventQueue) Push(evt interface{}) {
    q.mu.Lock()
    defer q.mu.Unlock()

    q.data.PushBack(evt)
}

func (q *eventQueue) Pop() interface{} {
    q.mu.Lock()
    defer q.mu.Unlock()

    front := q.data.Front()
    if front == nil {
        return nil
    }
    return q.data.Remove(front)
}

type eventQueueMap map[uint32]*eventQueue

func (m eventQueueMap) Push(mu *sync.Mutex, evt interface{}) {
    mu.Lock()
    defer mu.Unlock()
    for _, queue := range m {
        queue.Push(evt)
    }
}

