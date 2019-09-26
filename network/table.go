package network

import (
	"sync"
)

type findNodesTable struct {
	items map[SessionID]chan *FindNodesResult
	sync.Mutex
}

func newFindNodesTable() *findNodesTable {
	return &findNodesTable{
		items: make(map[SessionID]chan *FindNodesResult),
		Mutex: sync.Mutex{},
	}
}

func (t *findNodesTable) Put(id SessionID, ch chan *FindNodesResult) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = ch
}

func (t *findNodesTable) Get(id SessionID) chan *FindNodesResult {
	t.Lock()
	defer t.Unlock()
	return t.items[id]
}

func (t *findNodesTable) Remove(id SessionID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}

type findValueTable struct {
	items map[SessionID]chan *FindValueResult
	sync.Mutex
}

func newFindValueTable() *findValueTable {
	return &findValueTable{
		items: make(map[SessionID]chan *FindValueResult),
		Mutex: sync.Mutex{},
	}
}

func (t *findValueTable) Put(id SessionID, ch chan *FindValueResult) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = ch
}

func (t *findValueTable) Get(id SessionID) chan *FindValueResult {
	t.Lock()
	defer t.Unlock()
	return t.items[id]
}

func (t *findValueTable) Remove(id SessionID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}

type pingTable struct {
	items map[SessionID]chan *PingResult
	sync.Mutex
}

func newPingTable() *pingTable {
	return &pingTable{
		items: make(map[SessionID]chan *PingResult),
		Mutex: sync.Mutex{},
	}
}

func (t *pingTable) Put(id SessionID, ch chan *PingResult) {
	t.Lock()
	defer t.Unlock()
	t.items[id] = ch
}

func (t *pingTable) Get(id SessionID) chan *PingResult {
	t.Lock()
	defer t.Unlock()
	return t.items[id]
}

func (t *pingTable) Remove(id SessionID) {
	t.Lock()
	defer t.Unlock()
	delete(t.items, id)
}
