package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-program/host/flags"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli/v2"
)

var (
	ErrMissingRollupConfig = errors.New("missing rollup config")
	ErrMissingL2Genesis    = errors.New("missing l2 genesis")
	ErrInvalidL1Head       = errors.New("invalid l1 head")
	ErrInvalidL2Head       = errors.New("invalid l2 head")
	ErrInvalidL2OutputRoot = errors.New("invalid l2 output root")
	ErrL1AndL2Inconsistent = errors.New("l1 and l2 options must be specified together or both omitted")
	ErrInvalidL2Claim      = errors.New("invalid l2 claim")
	ErrInvalidL2ClaimBlock = errors.New("invalid l2 claim block number")
	ErrDataDirRequired     = errors.New("datadir must be specified when in non-fetching mode")
	ErrNoExecInServerMode  = errors.New("exec command must not be set when in server mode")
)

type Config struct {
	// DataDir is the directory to read/write pre-image data from/to.
	// If not set, an in-memory key-value store is used and fetching data must be enabled
	DataDir string

	// L1Head is the block hash of the L1 chain head block
	L1Head      common.Hash
	L1URL       string
	L1BeaconURL string
	L1TrustRPC  bool
	L1RPCKind   sources.RPCProviderKind

	// ExecCmd specifies the client program to execute in a separate process.
	// If unset, the fault proof client is run in the same process.
	ExecCmd string

	// ServerMode indicates that the program should run in pre-image server mode and wait for requests.
	// No client program is run.
	ServerMode bool

	// IsCustomChainConfig indicates that the program uses a custom chain configuration
	IsCustomChainConfig bool
}

func (c *Config) Check() error {
	if c.L1Head == (common.Hash{}) {
		return ErrInvalidL1Head
	}
	if !c.FetchingEnabled() && c.DataDir == "" {
		return ErrDataDirRequired
	}
	if c.ServerMode && c.ExecCmd != "" {
		return ErrNoExecInServerMode
	}
	return nil
}

func (c *Config) FetchingEnabled() bool {
	// TODO: Include Beacon URL once cancun is active on all chains we fault prove.
	return c.L1URL != ""
}

// NewConfig creates a Config with all optional values set to the CLI default value
func NewConfig(
	l1Head common.Hash,
) *Config {
	return &Config{
		L1Head:              l1Head,
		L1RPCKind:           sources.RPCKindStandard,
		IsCustomChainConfig: false,
	}
}

func NewConfigFromCLI(log log.Logger, ctx *cli.Context) (*Config, error) {
	if err := flags.CheckRequired(ctx); err != nil {
		return nil, err
	}
	l1Head := common.HexToHash(ctx.String(flags.L1Head.Name))
	if l1Head == (common.Hash{}) {
		return nil, ErrInvalidL1Head
	}
	return &Config{
		DataDir:             ctx.String(flags.DataDir.Name),
		L1Head:              l1Head,
		L1URL:               ctx.String(flags.L1NodeAddr.Name),
		L1BeaconURL:         ctx.String(flags.L1BeaconAddr.Name),
		L1TrustRPC:          ctx.Bool(flags.L1TrustRPC.Name),
		L1RPCKind:           sources.RPCProviderKind(ctx.String(flags.L1RPCProviderKind.Name)),
		ExecCmd:             ctx.String(flags.Exec.Name),
		ServerMode:          ctx.Bool(flags.Server.Name),
		IsCustomChainConfig: false,
	}, nil
}

func loadChainConfigFromGenesis(path string) (*params.ChainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read l2 genesis file: %w", err)
	}
	var genesis core.Genesis
	err = json.Unmarshal(data, &genesis)
	if err != nil {
		return nil, fmt.Errorf("parse l2 genesis file: %w", err)
	}
	return genesis.Config, nil
}
