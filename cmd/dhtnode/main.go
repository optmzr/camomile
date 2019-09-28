package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"

	"github.com/optmzr/d7024e-dht/ctl"
	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

func rpcServe(dht *dht.DHT) {
	api := ctl.NewAPI(dht)

	err := rpc.Register(api)
	if err != nil {
		log.Fatalln(err)
	}

	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	err = http.Serve(l, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func flagSplit(flag string) (string, string) {
	if flag == "" {
		return "", ""
	}

	components := strings.Split(flag, "@")
	nodeID := components[0]
	address := components[1]

	return nodeID, address
}

func main() {
	address, err := net.ResolveUDPAddr("udp", network.UdpPort)
	if err != nil {
		log.Fatalln(err)
	}

	meFlag := flag.String("me", "", "Defaults to an autogenerated ID, ip defaults to localhost")
	otherFlag := flag.String("other", "", "Waits for incoming connections if not supplied")

	flag.Parse()

	var others []route.Contact

	otherID, otherAddress := flagSplit(*otherFlag)
	if (otherID == "") || (otherAddress == "") {
		fmt.Println("asss")
		others = []route.Contact{
			route.Contact{
				NodeID:  node.NewID(),
				Address: *address,
			},
		}
	} else {
		nodeID, err := node.IDFromString(otherID)
		if err != nil {
			log.Fatalln(err)
		}

		nodeAddress, err := net.ResolveUDPAddr("udp", otherAddress)
		if err != nil {
			log.Fatalln(err)
		}

		others = []route.Contact{
			route.Contact{
				NodeID:  nodeID,
				Address: *nodeAddress,
			},
		}
	}

	var me route.Contact

	meID, meAddress := flagSplit(*meFlag)
	if (meID == "") || (meAddress == "") {
		me = route.Contact{
			NodeID:  node.NewID(),
			Address: *address,
		}
	} else {
		nodeID, err := node.IDFromString(meID)
		if err != nil {
			log.Fatalln(err)
		}

		nodeAddress, err := net.ResolveUDPAddr("udp", meAddress)
		if err != nil {
			log.Fatalln(err)
		}

		me = route.Contact{
			NodeID:  nodeID,
			Address: *nodeAddress,
		}
	}

	log.Printf("My node ID is: %v", me.NodeID)

	nw, err := network.NewUDPNetwork(me)
	if err != nil {
		log.Fatalln(err)
	}

	dht, err := dht.New(me, others, nw)
	if err != nil {
		log.Fatalln(err)
	}

	go rpcServe(dht)

	err = nw.Listen()
	if err != nil {
		log.Fatalln(err)
	}
}
