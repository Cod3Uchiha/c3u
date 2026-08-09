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
        assertEq(token.totalMinted(), 20_062_709 * 1e8);
        assertEq(token.balanceOf(owner), token.GENESIS_SUPPLY());
    }

    function testOwnerCanMintOnlyInsideHardCap() public {
        vm.prank(owner);
        token.mintRemaining(alice, 100 * 1e8);
        assertEq(token.balanceOf(alice), 100 * 1e8);
        assertEq(token.totalMinted(), token.GENESIS_SUPPLY() + 100 * 1e8);
    }

    function testNonOwnerCannotMint() public {
        vm.prank(alice);
        vm.expectRevert();
        token.mintRemaining(alice, 1e8);
    }

    function testCannotEverExceedTwentyOneMillion() public {
        uint256 tooMuch = token.remainingMintableSupply() + 1;
        vm.prank(owner);
        vm.expectRevert();
        token.mintRemaining(alice, tooMuch);
    }

    function testCanMintExactlyToCap() public {
        uint256 remaining = token.remainingMintableSupply();
        vm.prank(owner);
        token.mintRemaining(alice, remaining);
        assertEq(token.totalMinted(), token.MAX_SUPPLY());
        assertEq(token.remainingMintableSupply(), 0);
    }

    function testHolderCanBurn() public {
        vm.prank(owner);
        token.transfer(alice, 10 * 1e8);
        vm.prank(alice);
        token.burn(4 * 1e8);
        assertEq(token.balanceOf(alice), 6 * 1e8);
        assertEq(token.totalMinted(), token.GENESIS_SUPPLY());
    }

    function testBurnDoesNotReopenLifetimeIssuance() public {
        uint256 remaining = token.remainingMintableSupply();
        vm.prank(owner);
        token.mintRemaining(owner, remaining);

        vm.prank(owner);
        token.burn(1_000 * 1e8);

        assertEq(token.totalMinted(), token.MAX_SUPPLY());
        assertEq(token.remainingMintableSupply(), 0);

        vm.prank(owner);
        vm.expectRevert();
        token.mintRemaining(alice, 1);
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
