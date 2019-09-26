package dht

import (
	"fmt"
	"log"
	"net"

	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
	"golang.org/x/crypto/blake2b"
)

const α = 3 // Degree of parallelism.

type DHT struct {
	rt *route.Table
	nw network.Network
}

func New(me route.Contact, others []route.Contact, nw network.Network) (dht *DHT, err error) {
	dht = new(DHT)
	dht.rt, err = route.NewTable(me, others)
	if err != nil {
		err = fmt.Errorf("cannot initialize routing table: %w", err)
		return
	}

	dht.nw = nw

	return
}

// Get retrieves the value for a specified key from the network.
func (dht *DHT) Get(hash store.Key) (value string, err error) {
	value, err = dht.iterativeFindValue(hash)
	return
}

// Put stores the provided value in the network and returns a key.
func (dht *DHT) Put(value string) (hash store.Key, err error) {
	hash, err = dht.iterativeStore(value)
	return
}

// Join initiates a node lookup of itself to bootstrap the node into the
// network.
func (dht *DHT) Join(me route.Contact) (err error) {
	call := NewFindNodesCall(me.NodeID)
	contacts, err := dht.walk(call)
	if err != nil {
		return
	}

	logAcquaintedWith(contacts)
	return
}

type Call interface {
	Do(nw network.Network, address net.UDPAddr) (ch chan network.Result, err error)
	Result(result network.Result) (done bool)
	Target() (target node.ID)
}

func NewFindNodesCall(target node.ID) *FindNodesCall {
	return &FindNodesCall{
		target: target,
	}
}

type FindNodesCall struct {
	target node.ID
}

func (q *FindNodesCall) Do(nw network.Network, address net.UDPAddr) (chan network.Result, error) {
	ch, err := nw.FindNodes(q.target, address)
	return ch, err
}

func (q *FindNodesCall) Result(result network.Result) (done bool) {
	res, ok := result.(*network.FindNodesResult)
	if !ok {
		fmt.Println("invalid", result)
		return false
	}
	fmt.Println("valid", res)
	return true
}

func (q *FindNodesCall) Target() node.ID {
	return q.target
}

func (dht *DHT) walk(call Call) ([]route.Contact, error) {
	nw := dht.nw
	rt := dht.rt
	target := call.Target()

	// The first α contacts selected are used to create a *shortlist* for the
	// search.
	sl := dht.rt.NClosest(target, α)

	// Keep a map of contacts that has been sent to, to make sure we do not
	// contact the same node multiple times.
	sent := make(map[node.ID]bool)

	// If a cycle results in an unchanged `closest` node, then a FindNode
	// network call should be made to each of the closest nodes that has not
	// already been queried.
	rest := false

	// Contacts holds a sorted (slice) copy of the shortlist.
	contacts := sl.SortedContacts()

	// Closest is the node that closest in distance to the target node ID.
	closest := contacts[0]

	for {
		// Holds a slice of channels that are awaiting a response from the
		// network.
		await := [](chan network.Result){}

		for i, contact := range contacts {
			if i >= α && !rest {
				break // Limit to α contacts per shortlist.
			}
			if sent[contact.NodeID] {
				continue // Ignore already contacted contacts.
			}

			ch, err := call.Do(nw, contact.Address)
			if err != nil {
				sl.Remove(contact)
			} else {
				// Mark as contacted.
				sent[contact.NodeID] = true

				// Add to await channel queue.
				await = append(await, ch)
			}
		}

		results := make(chan network.Result)
		for _, ch := range await {
			// TODO(optmzr): Should this handle timeouts, or the network
			// package?
			go func(ch chan network.Result) {
				// Redirect all responses to the results channel.
				r := <-ch
				results <- r
			}(ch)
		}

		// Iterate through every result from the responding nodes and add their
		// closest contacts to the shortlist.
		for i := 0; i < len(await); i++ {
			// TODO(optmzr): Should this handle timeouts, or the network
			// package?
			result := <-results
			if result != nil {
				// Add node so it is moved to the top of its bucket in the
				// routing table.
				rt.Add(result.From())

				// Add the responding node's closest contacts.
				sl.Add(result.Closest()...)
			} else {
				// Network call timed out. Remove the callee from the shortlist.
				sl.Remove(result.From())
			}
		}

		contacts = sl.SortedContacts()
		first := contacts[0]
		if closest.NodeID.Equal(first.NodeID) {
			// Unchanged closest node from last run, re-run but check all the
			// nodes in the shortlist (and not only the α closest).
			if !rest {
				rest = true
				continue
			}

			// Done. Return the contacts in the shortlist sorted by distance.
			return contacts, nil

		} else {
			// New closest node found, continue iteration.
			closest = first
		}
	}
}

func (dht *DHT) iterativeStore(value string) (hash store.Key, err error) {
	hash = blake2b.Sum256([]byte(value))

	call := NewFindNodesCall(node.ID(hash))
	contacts, err := dht.walk(call)
	if err != nil {
		return
	}

	var stored []route.Contact
	for _, contact := range contacts {
		if e := dht.nw.Store(hash, value, contact.Address); e != nil {
			log.Printf("Failed to store at %s (%s): %v",
				contact.NodeID.String(), contact.Address.String(), e)
		} else {
			stored = append(stored, contact)
		}
	}

	if len(stored) > 0 {
		logStoredAt(hash, stored)
	}

	return
}

func (dht *DHT) iterativeFindValue(hash store.Key) (value string, err error) {
	return "", nil
}

func logStoredAt(hash store.Key, contacts []route.Contact) {
	log.Printf("Stored value with hash %v at %d nodes:\n",
		hash.String(), len(contacts))
	logContacts(contacts)
}

func logAcquaintedWith(contacts []route.Contact) {
	log.Printf("Acquainted with %d nodes:\n", len(contacts))
	logContacts(contacts)
}

func logContacts(contacts []route.Contact) {
	for _, contact := range contacts {
		log.Println("\t", contact.NodeID.String())
	}
}
