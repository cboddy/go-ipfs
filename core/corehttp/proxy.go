package corehttp

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"

	//version "github.com/ipfs/go-ipfs"
	core "github.com/ipfs/go-ipfs/core"
	p2p "github.com/ipfs/go-ipfs/p2p"
	peer "gx/ipfs/QmcZSzKEM5yDfpZbeEEZaVmaZ1zXm6JWTbrQZSB8hCVPzk/go-libp2p-peer"
	//coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	//id "gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/protocol/identify"
)

func ProxyOption() ServeOption {
	return func(ipfsNode *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.HandleFunc("/proxy/", func(w http.ResponseWriter, request *http.Request) {
			//get free tcp port
			_p2p := ipfsNode.P2P

			// parse request
			parsedRequest, err := parseRequest(request)
			if err != nil {
				// TODO: send error response
				fmt.Println(err)
				return
			}
			fmt.Println("parsed proxy req")
			fmt.Println(parsedRequest)

			stream, err := _p2p.NewStreamTo(ipfsNode.Context(), parsedRequest.target, "/p2p/"+parsedRequest.name)
			fmt.Printf("OPENED STREAM %s\n", stream)
			if err != nil {
				fmt.Println(err)
				return
			}
			// serialize proxy request
			proxyReq, err := http.NewRequest(request.Method, parsedRequest.httpPath, request.Body)
			if err != nil {
				fmt.Println(err)
				// TODO: send error response
				return
			}
			// send request to proxy target
			s := bufio.NewReader(stream)

			fmt.Printf("OPENED READER TO STREAM %s\n", proxyReq)
			proxyResponse, err := http.ReadResponse(s, proxyReq)
			fmt.Printf("GOT PROXY RESPONSE %s\n", proxyResponse)
			defer func() { proxyResponse.Body.Close() }()
			if err != nil {
				// TODO: send error response
				fmt.Println(err)
				return
			}
			// send client response
			proxyResponse.Write(w)

			fmt.Println("Finished!")

		})
		return mux, nil
	}
}

type proxyRequest struct {
	target   peer.ID
	name     string
	httpPath string // path to send to the proxy-host
}

//from the url path parse the peer-ID, name and http path
// /http/$peer_id/$name/$http_path
func parseRequest(request *http.Request) (*proxyRequest, error) {
	path := request.URL.Path

	split := strings.SplitN(path, "/", 6)
	if split[2] != "http" {
		return nil, fmt.Errorf("Invalid proxy request protocol '%s'", path)
	}

	if len(split) < 6 {
		return nil, fmt.Errorf("Invalid request path '%s'", path)
	}

	peerID, err := peer.IDB58Decode(split[3])

	if err != nil {
		return nil, err
	}

	return &proxyRequest{peerID, split[4], split[5]}, nil
}

// get stream from p2p registry or nil if not present
func getStreamForPeer(ipfsNode *core.IpfsNode, peerID *peer.ID) *p2p.StreamInfo {
	for _, stream := range ipfsNode.P2P.Streams.Streams {
		//TODO: is this comparing?
		fmt.Printf("Hey! found peer %s", stream.RemotePeer)
		if stream.RemotePeer == *peerID {
			return stream
		}
	}
	return nil
}
