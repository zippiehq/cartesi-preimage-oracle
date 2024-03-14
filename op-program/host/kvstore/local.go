package kvstore

import (

	"github.com/ethereum-optimism/optimism/op-program/client"
	"github.com/ethereum-optimism/optimism/op-program/host/config"
	"github.com/ethereum/go-ethereum/common"
)

type LocalPreimageSource struct {
	config *config.Config
}

func NewLocalPreimageSource(config *config.Config) *LocalPreimageSource {
	return &LocalPreimageSource{config}
}

var (
	l1HeadKey             = client.L1HeadLocalIndex.PreimageKey()
)

func (s *LocalPreimageSource) Get(key common.Hash) ([]byte, error) {
	switch [32]byte(key) {
	case l1HeadKey:
		return s.config.L1Head.Bytes(), nil
	default:
		return nil, ErrNotFound
	}
}
