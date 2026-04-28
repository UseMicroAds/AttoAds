// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import {AttoAdsEscrow} from "../src/AttoAdsEscrow.sol";
import {MockUSDC} from "./mocks/MockUSDC.sol";

contract AttoAdsEscrowTest is Test {
    AttoAdsEscrow public escrow;
    MockUSDC public usdc;

    address owner = address(this);
    address operator = makeAddr("operator");
    address advertiser = makeAddr("advertiser");
    address commenter = makeAddr("commenter");

    bytes32 campaignId = keccak256("campaign-1");
    bytes32 dealId = keccak256("deal-1");

    function setUp() public {
        usdc = new MockUSDC();
        escrow = new AttoAdsEscrow(address(usdc), operator);

        usdc.mint(advertiser, 1000e6);
    }

    function test_deposit() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 500e6);
        escrow.deposit(campaignId, 500e6);
        vm.stopPrank();

        assertEq(escrow.getAvailableBalance(campaignId), 500e6);
        assertEq(usdc.balanceOf(address(escrow)), 500e6);
    }

    function test_release() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 500e6);
        escrow.deposit(campaignId, 500e6);
        vm.stopPrank();

        vm.prank(operator);
        escrow.release(dealId, campaignId, commenter, 35e6);

        assertEq(usdc.balanceOf(commenter), 35e6);
        assertEq(escrow.getAvailableBalance(campaignId), 465e6);
    }

    function test_release_reverts_not_operator() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 500e6);
        escrow.deposit(campaignId, 500e6);
        vm.stopPrank();

        vm.prank(advertiser);
        vm.expectRevert(AttoAdsEscrow.NotOperator.selector);
        escrow.release(dealId, campaignId, commenter, 35e6);
    }

    function test_release_reverts_insufficient() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 50e6);
        escrow.deposit(campaignId, 50e6);
        vm.stopPrank();

        vm.prank(operator);
        vm.expectRevert(AttoAdsEscrow.InsufficientEscrow.selector);
        escrow.release(dealId, campaignId, commenter, 100e6);
    }

    function test_refund() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 500e6);
        escrow.deposit(campaignId, 500e6);
        vm.stopPrank();

        vm.prank(operator);
        escrow.release(dealId, campaignId, commenter, 100e6);

        vm.prank(advertiser);
        escrow.refund(campaignId);

        assertEq(usdc.balanceOf(advertiser), 900e6);
        assertEq(escrow.getAvailableBalance(campaignId), 0);
    }

    function test_refund_reverts_not_advertiser() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 500e6);
        escrow.deposit(campaignId, 500e6);
        vm.stopPrank();

        vm.prank(commenter);
        vm.expectRevert(AttoAdsEscrow.NotAdvertiser.selector);
        escrow.refund(campaignId);
    }

    function test_refund_reverts_already_refunded() public {
        vm.startPrank(advertiser);
        usdc.approve(address(escrow), 500e6);
        escrow.deposit(campaignId, 500e6);
        vm.stopPrank();

        vm.prank(advertiser);
        escrow.refund(campaignId);

        vm.prank(advertiser);
        vm.expectRevert(AttoAdsEscrow.AlreadyRefunded.selector);
        escrow.refund(campaignId);
    }

    function test_set_operator() public {
        address newOp = makeAddr("newOperator");
        escrow.setOperator(newOp);
        assertEq(escrow.operator(), newOp);
    }

    function test_set_operator_reverts_not_owner() public {
        vm.prank(advertiser);
        vm.expectRevert(AttoAdsEscrow.NotOwner.selector);
        escrow.setOperator(advertiser);
    }
}
