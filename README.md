# C3U

C3U is a standalone capped cryptocurrency token designed for the Base network.

## Token specification

- Name: C3U
- Symbol: C3U
- Standard: ERC-20
- Network target: Base
- Decimals: 8
- Genesis supply: 20,062,709 C3U
- Maximum supply: 21,000,000 C3U
- Burnable: yes
- Transfer tax: none
- Blacklist: none
- Rebase: none
- Hard cap enforced by the smart contract
- Remaining issuance can only occur within the 21,000,000 C3U lifetime cap
- Two-step ownership transfer

The genesis supply is a Bitcoin circulating-supply snapshot selected for the C3U launch design on 2026-08-09. C3U is an independent cryptocurrency project; matching Bitcoin's supply structure does not make C3U equal in price or value to BTC.

## Security model

C3U uses OpenZeppelin `ERC20Capped`, `ERC20Burnable`, and `Ownable2Step`. The token contract cannot mint beyond 21,000,000 C3U.

Before any mainnet deployment, test the contract on Base Sepolia and review the final source and ownership setup.

## Development

Requires Foundry.

```bash
forge install --no-git OpenZeppelin/openzeppelin-contracts@v5.6.1
forge install --no-git foundry-rs/forge-std
forge build
forge test -vv
```

## Base Sepolia deployment

Import a deployment wallet into Foundry's encrypted keystore. Never commit a seed phrase or private key.

```bash
cast wallet import c3u-deployer --interactive
export C3U_OWNER=0xYOUR_PUBLIC_OWNER_WALLET
export BASE_SEPOLIA_RPC_URL=https://sepolia.base.org

forge script script/DeployC3U.s.sol:DeployC3U \
  --rpc-url base_sepolia \
  --account c3u-deployer \
  --broadcast
```

After deployment, verify the contract address and test transfers, burns, ownership transfer, and capped issuance before considering Base mainnet.
