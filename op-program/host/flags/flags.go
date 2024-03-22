package flags

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	service "github.com/ethereum-optimism/optimism/op-service"
	openum "github.com/ethereum-optimism/optimism/op-service/enum"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

const EnvVarPrefix = "OP_PROGRAM"

func prefixEnvVars(name string) []string {
	return service.PrefixEnvVar(EnvVarPrefix, name)
}

var (
	Network = &cli.StringFlag{
		Name:    "network",
		Usage:   fmt.Sprintf("Predefined network selection. Available networks: %s", strings.Join(chaincfg.AvailableNetworks(), ", ")),
		EnvVars: prefixEnvVars("NETWORK"),
	}
	DataDir = &cli.StringFlag{
		Name:    "datadir",
		Usage:   "Directory to use for preimage data storage. Default uses in-memory storage",
		EnvVars: prefixEnvVars("DATADIR"),
	}
	L1Head = &cli.StringFlag{
		Name:    "l1.head",
		Usage:   "Hash of the L1 head block. Derivation stops after this block is processed.",
		EnvVars: prefixEnvVars("L1_HEAD"),
	}
	L1NodeAddr = &cli.StringFlag{
		Name:    "l1",
		Usage:   "Address of L1 JSON-RPC endpoint to use (eth namespace required)",
		EnvVars: prefixEnvVars("L1_RPC"),
	}
	L1BeaconAddr = &cli.StringFlag{
		Name:    "l1.beacon",
		Usage:   "Address of L1 Beacon API endpoint to use",
		EnvVars: prefixEnvVars("L1_BEACON_API"),
	}
	L1TrustRPC = &cli.BoolFlag{
		Name:    "l1.trustrpc",
		Usage:   "Trust the L1 RPC, sync faster at risk of malicious/buggy RPC providing bad or inconsistent L1 data",
		EnvVars: prefixEnvVars("L1_TRUST_RPC"),
	}
	L1RPCProviderKind = &cli.GenericFlag{
		Name: "l1.rpckind",
		Usage: "The kind of RPC provider, used to inform optimal transactions receipts fetching, and thus reduce costs. Valid options: " +
			openum.EnumString(sources.RPCProviderKinds),
		EnvVars: prefixEnvVars("L1_RPC_KIND"),
		Value: func() *sources.RPCProviderKind {
			out := sources.RPCKindStandard
			return &out
		}(),
	}
	Exec = &cli.StringFlag{
		Name:    "exec",
		Usage:   "Run the specified client program as a separate process detached from the host. Default is to run the client program in the host process.",
		EnvVars: prefixEnvVars("EXEC"),
	}
	Server = &cli.BoolFlag{
		Name:    "server",
		Usage:   "Run in pre-image server mode without executing any client program.",
		EnvVars: prefixEnvVars("SERVER"),
	}
	APIAddress = &cli.StringFlag{
		Name:    "api.address",
		Usage:   "Http API address.",
		EnvVars: prefixEnvVars("API_ADDRESS"),
	}
)

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

var requiredFlags = []cli.Flag{
	L1Head,
}

var programFlags = []cli.Flag{
	Network,
	DataDir,
	L1NodeAddr,
	L1BeaconAddr,
	L1TrustRPC,
	L1RPCProviderKind,
	Exec,
	Server,
	APIAddress,
}

func init() {
	Flags = append(Flags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags, requiredFlags...)
	Flags = append(Flags, programFlags...)
}

func CheckRequired(ctx *cli.Context) error {
	for _, flag := range requiredFlags {
		if !ctx.IsSet(flag.Names()[0]) {
			return fmt.Errorf("flag %s is required", flag.Names()[0])
		}
	}
	return nil
}
