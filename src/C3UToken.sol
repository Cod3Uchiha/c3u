// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC20Burnable} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import {ERC20Capped} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Capped.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Ownable2Step} from "@openzeppelin/contracts/access/Ownable2Step.sol";

/// @title C3U
/// @notice Bitcoin-style capped ERC-20 cryptocurrency for the Base network.
/// @dev C3U uses 8 decimals to mirror Bitcoin's display precision. There are no taxes,
///      blacklists, rebases, transfer fees, or hidden mint paths. Supply can never exceed 21M C3U.
contract C3UToken is ERC20, ERC20Burnable, ERC20Capped, Ownable2Step {
    uint8 public constant C3U_DECIMALS = 8;
    uint256 public constant UNIT = 10 ** C3U_DECIMALS;

    /// @notice Absolute lifetime supply ceiling: 21 million C3U.
    uint256 public constant MAX_SUPPLY = 21_000_000 * UNIT;

    /// @notice Launch supply snapshot matching the BTC circulating-supply figure selected on 2026-08-09.
    uint256 public constant GENESIS_SUPPLY = 20_062_709 * UNIT;

    /// @notice Emitted whenever additional C3U is created from the remaining capped supply.
    event EmissionMinted(address indexed to, uint256 amount, uint256 totalSupplyAfter);

    constructor(address initialOwner) ERC20("C3U", "C3U") ERC20Capped(MAX_SUPPLY) Ownable(initialOwner) {
        _mint(initialOwner, GENESIS_SUPPLY);
    }

    /// @notice C3U intentionally uses eight decimals, like BTC.
    function decimals() public pure override returns (uint8) {
        return C3U_DECIMALS;
    }

    /// @notice Amount that can still ever be minted before the hard 21M cap is reached.
    function remainingMintableSupply() public view returns (uint256) {
        return cap() - totalSupply();
    }

    /// @notice Mint from the permanently limited remainder between current supply and 21M.
    /// @dev Only the owner can emit new C3U. ERC20Capped makes exceeding 21M impossible.
    function mintRemaining(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
        emit EmissionMinted(to, amount, totalSupply());
    }

    /// @dev Required by Solidity because ERC20Capped overrides ERC20._update.
    function _update(address from, address to, uint256 value) internal override(ERC20, ERC20Capped) {
        super._update(from, to, value);
    }
}
