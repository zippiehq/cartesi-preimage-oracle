package host

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	"github.com/ethereum-optimism/optimism/op-program/client/l1"
	"github.com/ethereum-optimism/optimism/op-program/host/config"
	"github.com/ethereum-optimism/optimism/op-program/host/flags"
	"github.com/ethereum-optimism/optimism/op-program/host/kvstore"
	"github.com/ethereum-optimism/optimism/op-program/host/prefetcher"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type L2Source struct {
	*L2Client
	*sources.DebugClient
}

func Main(logger log.Logger, cfg *config.Config) error {
	if err := cfg.Check(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, logger)

	return PreimageServer(context.Background(), logger, cfg)
}

// PreimageServer reads hints and preimage requests from the provided channels and processes those requests.
// This method will block until both the hinter and preimage handlers complete.
// If either returns an error both handlers are stopped.
// The supplied preimageChannel and hintChannel will be closed before this function returns.
func PreimageServer(ctx context.Context, logger log.Logger, cfg *config.Config) error {
	logger.Info("Starting preimage server")

	var kv kvstore.KV
	if cfg.DataDir == "" {
		logger.Info("Using in-memory storage")
		kv = kvstore.NewMemKV()
	} else {
		logger.Info("Creating disk storage", "datadir", cfg.DataDir)
		if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
			return fmt.Errorf("creating datadir: %w", err)
		}
		kv = kvstore.NewDiskKV(cfg.DataDir)
	}

	var (
		preimageSource kvstore.PreimageSource
		hintHander     preimage.HintHandler
	)
	if cfg.FetchingEnabled() {
		prefetch, err := makePrefetcher(ctx, logger, kv, cfg)
		if err != nil {
			return fmt.Errorf("failed to create prefetcher: %w", err)
		}
		preimageSource = func(key common.Hash) ([]byte, error) { return prefetch.GetPreimage(ctx, key) }
		hintHander = prefetch.Hint
	} else {
		logger.Info("Using offline mode. All required pre-images must be pre-populated.")
		preimageSource = kv.Get
		hintHander = func(hint string) error {
			logger.Debug("ignoring prefetch hint", "hint", hint)
			return nil
		}
	}

	return httpServer(logger, cfg.APIAddress, preimageSource, hintHander)
}

func makePrefetcher(ctx context.Context, logger log.Logger, kv kvstore.KV, cfg *config.Config) (*prefetcher.Prefetcher, error) {
	logger.Info("Connecting to L1 node", "l1", cfg.L1URL)
	l1RPC, err := client.NewRPC(ctx, logger, cfg.L1URL, client.WithDialBackoff(10))
	if err != nil {
		return nil, fmt.Errorf("failed to setup L1 RPC: %w", err)
	}

	l1ClCfg := sources.L1ClientDefaultConfig(cfg.L1TrustRPC, cfg.L1RPCKind)
	l1Cl, err := sources.NewL1Client(l1RPC, logger, nil, l1ClCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 client: %w", err)
	}
	l1Beacon := sources.NewBeaconHTTPClient(client.NewBasicHTTPClient(cfg.L1BeaconURL, logger))
	l1BlobFetcher := sources.NewL1BeaconClient(l1Beacon, sources.L1BeaconClientConfig{FetchAllSidecars: false})
	return prefetcher.NewPrefetcher(logger, l1Cl, l1BlobFetcher, kv), nil
}

func httpServer(
	logger log.Logger,
	hostPort string,
	preimageSource kvstore.PreimageSource,
	hintHandler preimage.HintHandler,
) error {
	http.HandleFunc("/dehash/", func(w http.ResponseWriter, req *http.Request) {
		keyStr := req.URL.Path[len("/dehash/"):]
		key, err := hex.DecodeString(keyStr)
		if err != nil {
			logger.Error("failed to decode key from hex", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		key[0] = 2 // keccak256

		val, err := preimageSource(common.Hash(key[:common.HashLength]))
		if err != nil {
			logger.Error("failed to get preimage value for key", keyStr, err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-type", "application/octet-stream")
			if _, err = w.Write(val); err != nil {
				logger.Error("failed to write preimage value to http response", err)
			}
		}
	})

	http.HandleFunc("/hint/", func(w http.ResponseWriter, req *http.Request) {
		hint := req.URL.Path[len("/hint/"):]

		if !strings.Contains(hint, l1.HintL1BlockHeader) &&
			!strings.Contains(hint, l1.HintL1Transactions) &&
			!strings.Contains(hint, l1.HintL1Receipts) &&
			!strings.Contains(hint, l1.HintL1Blob) &&
			!strings.Contains(hint, l1.HintL1KZGPointEvaluation) {
			logger.Error("invalid hint type")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := hintHandler(hint); err != nil {
			logger.Error("failed to process hint", err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-type", "application/octet-stream")
			w.Write([]byte("ok"))
		}
	})

	return http.ListenAndServe(hostPort, nil)
}
