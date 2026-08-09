// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Script} from "forge-std/Script.sol";
import {C3UToken} from "../src/C3UToken.sol";

contract DeployC3U is Script {
    function run() external returns (C3UToken token) {
        address owner = vm.envAddress("C3U_OWNER");
        vm.startBroadcast();
        token = new C3UToken(owner);
        vm.stopBroadcast();
    }
}
