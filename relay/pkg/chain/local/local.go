package local

import (
	"github.com/ipfs/go-log"
	"github.com/keep-network/tbtc/relay/pkg/chain"
)

var logger = log.Logger("relay-chain-local")

// localChain is a local implementation of the host chain interface.
type localChain struct{}

// Connect performs initialization for communication with the local blockchain.
func Connect() (chain.Handle, error) {
	logger.Infof("connecting local host chain")

	return &localChain{}, nil
}

// GetBestKnownDigest returns the best known digest.
func (lc *localChain) GetBestKnownDigest() ([32]uint8, error) {
	panic("not implemented yet")
}
