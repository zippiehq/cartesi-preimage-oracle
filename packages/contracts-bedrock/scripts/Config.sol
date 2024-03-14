// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Vm } from "forge-std/Vm.sol";

/// @title Config
/// @notice Contains all env var based config. Add any new env var parsing to this file
///         to ensure that all config is in a single place.
library Config {
    /// @notice Foundry cheatcode VM.
    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    /// @notice Returns the path on the local filesystem where the deployment artifact is
    ///         written to disk after doing a deployment.
    function deployArtifactPath(string memory _deploymentsDir) internal view returns (string memory _env) {
        _env = vm.envOr("DEPLOYMENT_OUTFILE", string.concat(_deploymentsDir, "/.deploy"));
    }

    /// @notice Returns the chainid from the EVM context or the value of the CHAIN_ID env var as
    ///         an override.
    function chainID() internal view returns (uint256 _env) {
        _env = vm.envOr("CHAIN_ID", block.chainid);
    }

    /// @notice Returns the value of the env var CONTRACT_ADDRESSES_PATH which is a JSON key/value
    ///         pair of contract names and their addresses. Each key/value pair is passed to `save`
    ///         which then backs the `getAddress` function.
    function contractAddressesPath() internal view returns (string memory _env) {
        _env = vm.envOr("CONTRACT_ADDRESSES_PATH", string(""));
    }

    /// @notice Returns the deployment context which was only useful in the hardhat deploy style
    ///         of deployments. It is now DEPRECATED and will be removed in the future.
    function deploymentContext() internal view returns (string memory _env) {
        _env = vm.envOr("DEPLOYMENT_CONTEXT", string(""));
    }

    /// @notice The CREATE2 salt to be used when deploying the implementations.
    function implSalt() internal view returns (string memory _env) {
        _env = vm.envOr("IMPL_SALT", string("ethers phoenix"));
    }

    /// @notice Returns the path that the state dump file should be written to or read from
    ///         on the local filesystem.
    function stateDumpPath(string memory _name) internal view returns (string memory _env) {
        _env = vm.envOr(
            "STATE_DUMP_PATH", string.concat(vm.projectRoot(), "/", _name, "-", vm.toString(block.chainid), ".json")
        );
    }

    /// @notice Returns the name of the script used for deployment. This was useful for parsing the
    ///         foundry deploy artifacts to turn them into hardhat deploy style artifacts. It is now
    ///         DEPRECATED and will be removed in the future.
    function deployScript(string memory _name) internal view returns (string memory _env) {
        _env = vm.envOr("DEPLOY_SCRIPT", _name);
    }

    /// @notice Returns the sig of the entrypoint to the deploy script. By default, it is `run`.
    ///         This was useful for creating hardhat deploy style artifacts and will be removed in a future release.
    function sig() internal view returns (string memory _env) {
        _env = vm.envOr("SIG", string("run"));
    }

    /// @notice Returns the name of the file that the forge deployment artifact is written to on the local
    ///         filesystem. By default, it is the name of the deploy script with the suffix `-latest.json`.
    ///         This was useful for creating hardhat deploy style artifacts and will be removed in a future release.
    function deployFile(string memory _sig) internal view returns (string memory _env) {
        _env = vm.envOr("DEPLOY_FILE", string.concat(_sig, "-latest.json"));
    }

    /// @notice Returns the private key that is used to configure drippie.
    function drippieOwnerPrivateKey() internal view returns (uint256 _env) {
        _env = vm.envUint("DRIPPIE_OWNER_PRIVATE_KEY");
    }
}
