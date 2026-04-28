// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Script, console} from "forge-std/Script.sol";
import {AttoAdsEscrow} from "../src/AttoAdsEscrow.sol";

contract DeployScript is Script {
    // USDC on Base Sepolia
    address constant USDC_BASE_SEPOLIA = 0x036CbD53842c5426634e7929541eC2318f3dCF7e;

    function run() public {
        address operator = vm.envAddress("OPERATOR_ADDRESS");

        vm.startBroadcast();
        AttoAdsEscrow escrow = new AttoAdsEscrow(USDC_BASE_SEPOLIA, operator);
        vm.stopBroadcast();

        console.log("AttoAdsEscrow deployed at:", address(escrow));
    }
}
