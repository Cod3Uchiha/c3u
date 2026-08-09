// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";
import {C3UToken} from "../src/C3UToken.sol";

contract C3UTokenTest is Test {
    C3UToken internal token;
    address internal owner = makeAddr("owner");
    address internal alice = makeAddr("alice");
    address internal bob = makeAddr("bob");

    function setUp() public {
        token = new C3UToken(owner);
    }

    function testMetadataAndBitcoinStyleSupply() public view {
        assertEq(token.name(), "C3U");
        assertEq(token.symbol(), "C3U");
        assertEq(token.decimals(), 8);
        assertEq(token.cap(), 21_000_000 * 1e8);
        assertEq(token.totalSupply(), 20_062_709 * 1e8);
        assertEq(token.balanceOf(owner), token.GENESIS_SUPPLY());
    }

    function testOwnerCanMintOnlyInsideHardCap() public {
        vm.prank(owner);
        token.mintRemaining(alice, 100 * 1e8);
        assertEq(token.balanceOf(alice), 100 * 1e8);
    }

    function testNonOwnerCannotMint() public {
        vm.prank(alice);
        vm.expectRevert();
        token.mintRemaining(alice, 1e8);
    }

    function testCannotEverExceedTwentyOneMillion() public {
        vm.prank(owner);
        vm.expectRevert();
        token.mintRemaining(alice, token.remainingMintableSupply() + 1);
    }

    function testCanMintExactlyToCap() public {
        vm.prank(owner);
        token.mintRemaining(alice, token.remainingMintableSupply());
        assertEq(token.totalSupply(), token.MAX_SUPPLY());
        assertEq(token.remainingMintableSupply(), 0);
    }

    function testHolderCanBurn() public {
        vm.prank(owner);
        token.transfer(alice, 10 * 1e8);
        vm.prank(alice);
        token.burn(4 * 1e8);
        assertEq(token.balanceOf(alice), 6 * 1e8);
    }

    function testOwnershipTransferIsTwoStep() public {
        vm.prank(owner);
        token.transferOwnership(bob);
        assertEq(token.pendingOwner(), bob);
        vm.prank(bob);
        token.acceptOwnership();
        assertEq(token.owner(), bob);
    }
}
