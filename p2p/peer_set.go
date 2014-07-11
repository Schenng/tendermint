package p2p

import (
	"sync"
)

/*
ReadOnlyPeerSet has a subset of the methods of PeerSet.
*/
type ReadOnlyPeerSet interface {
	Has(addr *NetAddress) bool
	List() []*Peer
}

//-----------------------------------------------------------------------------

/*
PeerSet is a special structure for keeping a table of peers.
Iteration over the peers is super fast and thread-safe.
*/
type PeerSet struct {
	mtx    sync.Mutex
	lookup map[string]*peerSetItem
	list   []*Peer
}

type peerSetItem struct {
	peer  *Peer
	index int
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		lookup: make(map[string]*peerSetItem),
		list:   make([]*Peer, 0, 256),
	}
}

func (ps *PeerSet) Add(peer *Peer) bool {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	addr := peer.RemoteAddress().String()
	if ps.lookup[addr] != nil {
		return false
	}
	index := len(ps.list)
	// Appending is safe even with other goroutines
	// iterating over the ps.list slice.
	ps.list = append(ps.list, peer)
	ps.lookup[addr] = &peerSetItem{peer, index}
	return true
}

func (ps *PeerSet) Has(addr *NetAddress) bool {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	_, ok := ps.lookup[addr.String()]
	return ok
}

func (ps *PeerSet) Remove(peer *Peer) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	addr := peer.RemoteAddress().String()
	item := ps.lookup[addr]
	if item == nil {
		return
	}
	index := item.index
	// Copy the list but without the last element.
	// (we must copy because we're mutating the list)
	newList := make([]*Peer, len(ps.list)-1)
	copy(newList, ps.list)
	// If it's the last peer, that's an easy special case.
	if index == len(ps.list)-1 {
		ps.list = newList
		return
	}
	// Move the last item from ps.list to "index" in list.
	lastPeer := ps.list[len(ps.list)-1]
	lastPeerAddr := lastPeer.RemoteAddress().String()
	lastPeerItem := ps.lookup[lastPeerAddr]
	newList[index] = lastPeer
	lastPeerItem.index = index
	ps.list = newList
	delete(ps.lookup, addr)
}

func (ps *PeerSet) Size() int {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return len(ps.list)
}

// threadsafe list of peers.
func (ps *PeerSet) List() []*Peer {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.list
}