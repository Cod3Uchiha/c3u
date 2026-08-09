# C3U

<p align="center">
  <img src="assets/c3u-logo.jpg" alt="C3U official logo" width="320" />
</p>

<p align="center"><strong>C3U — a standalone, Bitcoin-style capped ERC-20 cryptocurrency for Base.</strong></p>

C3U is an independent cryptocurrency project. It is not linked to the Cod3Uchiha API platform.

## Status

**Base Sepolia testnet: deployed, source verified, transfer tested.**

C3U is not yet deployed to Base mainnet. The Base Sepolia contract exists for public testing before any production launch.

## Official Base Sepolia deployment

| Item | Value |
| --- | --- |
| Network | Base Sepolia |
| Chain ID | `84532` |
| Contract | `0x47407E17bf4915b280a29AEF9e8482ae3a88Df29` |
| Deployment transaction | `0xe2be2d87b2820aad546dc7ab76c8be557d825203852571154f79e97f1e2eb337` |
| Source | Verified on BaseScan |
| Owner/deployer | `0x42337F5384c4dF9617309dfEc36b25ea6d2f6f7b` |

Explorer:
- https://sepolia.basescan.org/address/0x47407E17bf4915b280a29AEF9e8482ae3a88Df29
- https://sepolia.basescan.org/tx/0xe2be2d87b2820aad546dc7ab76c8be557d825203852571154f79e97f1e2eb337

## Token specification

- Name: **C3U**
- Symbol: **C3U**
- Standard: **ERC-20**
- Target network: **Base**
- Decimals: **8**
- Genesis supply: **20,062,709 C3U**
- Maximum lifetime issuance: **21,000,000 C3U**
- Remaining issuance after genesis: **937,291 C3U**
- Burnable: **yes**
- Transfer tax: **none**
- Blacklist: **none**
- Rebase: **none**
- Two-step ownership transfer: **yes**

The genesis supply is a Bitcoin circulating-supply snapshot selected for the C3U launch design on 2026-08-09. Matching Bitcoin's numerical supply structure does **not** make C3U equal in price, market value, or backing to BTC.

## Supply rules

C3U has a permanent lifetime issuance ceiling of **21,000,000 C3U**. The contract tracks cumulative minted supply separately from current circulating supply. Burning tokens lowers `totalSupply()` but does not reopen minting capacity. This prevents burned C3U from being re-minted later and keeps total lifetime issuance at or below 21 million.

The owner can only mint from the unissued remainder. Once cumulative minted supply reaches 21,000,000 C3U, no additional C3U can ever be created by the contract.

## Official logo

The canonical C3U logo is [`assets/c3u-logo.jpg`](assets/c3u-logo.jpg), supplied by the project owner. Metadata integrations should use the canonical raw asset:

`https://raw.githubusercontent.com/Cod3Uchiha/c3u/main/assets/c3u-logo.jpg`

## Development

C3U uses Foundry and OpenZeppelin Contracts.

```bash
forge install --no-git OpenZeppelin/openzeppelin-contracts@v5.6.1
forge install --no-git foundry-rs/forge-std
forge fmt --check
forge build
forge test -vv
```

The test suite covers metadata, genesis supply, owner-only issuance, the 21M lifetime cap, burns, burn/mint-cap interaction, and two-step ownership transfer.

## Test the deployed token

```bash
export C3U_ADDRESS=0x47407E17bf4915b280a29AEF9e8482ae3a88Df29
export BASE_SEPOLIA_RPC_URL=https://sepolia.base.org

cast call "$C3U_ADDRESS" "name()(string)" --rpc-url "$BASE_SEPOLIA_RPC_URL"
cast call "$C3U_ADDRESS" "symbol()(string)" --rpc-url "$BASE_SEPOLIA_RPC_URL"
cast call "$C3U_ADDRESS" "decimals()(uint8)" --rpc-url "$BASE_SEPOLIA_RPC_URL"
cast call "$C3U_ADDRESS" "totalSupply()(uint256)" --rpc-url "$BASE_SEPOLIA_RPC_URL"
cast call "$C3U_ADDRESS" "remainingMintableSupply()(uint256)" --rpc-url "$BASE_SEPOLIA_RPC_URL"
```

## Deploy your own test instance

Never commit private keys or seed phrases. Import an account into Foundry's encrypted keystore:

```bash
cast wallet import c3u-deployer --interactive
export C3U_OWNER=0xYOUR_PUBLIC_OWNER_ADDRESS
export ETH_RPC_URL=https://sepolia.base.org

forge create src/C3UToken.sol:C3UToken \
  --rpc-url "$ETH_RPC_URL" \
  --account c3u-deployer \
  --broadcast \
  --constructor-args "$C3U_OWNER"
```

## Repository layout

```text
src/C3UToken.sol             C3U ERC-20 contract
script/DeployC3U.s.sol       Foundry deployment script
test/C3UToken.t.sol          contract tests
deployments/                 canonical deployment records
assets/c3u-logo.jpg          official C3U logo
token-metadata.json          machine-readable token metadata
SECURITY.md                  security and key-management notes
```

## Mainnet policy

The Base Sepolia deployment is the test version. A future Base mainnet deployment should only use a reviewed source commit and a deliberately selected owner wallet. The production contract address must be added to this repository after deployment so users can distinguish the canonical token from copies.

## License

MIT. See [`LICENSE`](LICENSE).
