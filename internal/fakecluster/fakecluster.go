package fakecluster

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/vipnode/vipnode/client"
	"github.com/vipnode/vipnode/host"
	"github.com/vipnode/vipnode/internal/fakenode"
	"github.com/vipnode/vipnode/jsonrpc2"
	"github.com/vipnode/vipnode/pool"
	"github.com/vipnode/vipnode/pool/store/memory"
)

type clusterHost struct {
	*host.Host
	Node *fakenode.FakeNode
	In   *jsonrpc2.Remote
	Out  *jsonrpc2.Remote
	Key  *ecdsa.PrivateKey
}

type clusterClient struct {
	*client.Client
	Node *fakenode.FakeNode
	In   *jsonrpc2.Remote
	Out  *jsonrpc2.Remote
	Key  *ecdsa.PrivateKey
}

// Cluster represents a set of active hosts and clients connected to a pool.
type Cluster struct {
	Clients []clusterClient
	Hosts   []clusterHost
	Pool    *pool.VipnodePool

	pipes []io.Closer
}

// New returns a pre-connected pool of hosts and clients.
func New(hostKeys []*ecdsa.PrivateKey, clientKeys []*ecdsa.PrivateKey) (*Cluster, error) {
	cluster := &Cluster{
		Hosts:   []clusterHost{},
		Clients: []clusterClient{},
		pipes:   []io.Closer{},
	}

	cluster.Pool = pool.New(memory.New(), nil)
	payout := ""
	for _, hostKey := range hostKeys {
		rpcPool2Host, rpcHost2Pool := jsonrpc2.ServePipe()
		cluster.pipes = append(cluster.pipes, rpcPool2Host, rpcHost2Pool)
		if err := rpcPool2Host.Server.Register("vipnode_", cluster.Pool); err != nil {
			return nil, err
		}

		hostNodeID := discv5.PubkeyID(&hostKey.PublicKey).String()
		hostNode := fakenode.Node(hostNodeID)
		hostNodeURI := fmt.Sprintf("enode://%s@127.0.0.1:30303", hostNodeID)
		h := host.New(hostNode, payout)
		if err := rpcHost2Pool.Server.RegisterMethod("vipnode_whitelist", h, "Whitelist"); err != nil {
			return nil, err
		}
		h.NodeURI = hostNodeURI
		hostPool := pool.Remote(rpcHost2Pool, hostKey)

		if err := h.Start(hostPool); err != nil {
			return nil, err
		}

		cluster.Hosts = append(cluster.Hosts, clusterHost{
			Host: h,
			Node: hostNode,
			In:   rpcPool2Host,
			Out:  rpcHost2Pool,
			Key:  hostKey,
		})
	}

	for _, clientKey := range clientKeys {
		rpcPool2Client, rpcClient2Pool := jsonrpc2.ServePipe()
		cluster.pipes = append(cluster.pipes, rpcPool2Client, rpcClient2Pool)
		rpcPool2Client.Server.Register("vipnode_", cluster.Pool)

		clientNodeID := discv5.PubkeyID(&clientKey.PublicKey).String()
		clientNode := fakenode.Node(clientNodeID)
		c := client.New(clientNode)
		clientPool := pool.Remote(rpcClient2Pool, clientKey)
		if err := c.Start(clientPool); err != nil {
			return nil, err
		}
		cluster.Clients = append(cluster.Clients, clusterClient{
			Client: c,
			Node:   clientNode,
			In:     rpcPool2Client,
			Out:    rpcClient2Pool,
			Key:    clientKey,
		})
	}
	return cluster, nil
}

// Close shuts down all the open pipes.
func (c *Cluster) Close() error {
	errors := []error{}
	for _, pipe := range c.pipes {
		if err := pipe.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	for _, host := range c.Hosts {
		host.Stop()
	}
	for _, client := range c.Clients {
		client.Stop()
	}
	for _, host := range c.Hosts {
		if err := host.Wait(); err != nil {
			errors = append(errors, err)
		}
	}
	for _, client := range c.Clients {
		if err := client.Wait(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return CloseErrors(errors)
	}
	return nil
}

// CloseErrors are used to return a set of errors that occurred while
// attempting to shut down a cluster.
type CloseErrors []error

func (e CloseErrors) Error() string {
	if len(e) == 0 {
		return "no close errors"
	}

	var s strings.Builder
	fmt.Fprintf(&s, "failed to close with %d errors: ", len(e))
	for i, err := range e {
		s.WriteString(err.Error())
		if i != len(e)-1 {
			s.WriteString("; ")
		}
	}
	return s.String()
}
