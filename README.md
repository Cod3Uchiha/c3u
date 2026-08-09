# C3U Core

<p align="center"><img src="assets/c3u-logo.jpg" width="260" alt="C3U logo"></p>

**C3U Core is the native C3U blockchain.** It does not require Ethereum, Base, ERC-20, a contract address, or ETH gas. C3U is the network's native currency and transaction fees are paid in C3U.

> Status: **experimental native-chain testnet/regtest implementation.** It is suitable for development and public testing, not yet a production-value mainnet.

## Monetary policy

- Native currency: **C3U**
- Smallest unit: **1 sat = 0.00000001 C3U**
- Precision: **8 decimals**
- Initial block subsidy: **50 C3U**
- Halving interval (mainnet/testnet): **210,000 blocks**
- Target block interval (mainnet): **10 minutes**
- Lifetime issuance ceiling: **21,000,000 C3U**
- Genesis premine: **none**
- Consensus: **Proof of Work**
- Ledger model: **UTXO**

Like Bitcoin's subsidy schedule, integer satoshi rounding means the final issued amount approaches but does not exceed 21 million.

## What is implemented

- deterministic C3U genesis block per network
- independent `mainnet`, `testnet`, and `regtest`
- native C3U block rewards
- Bitcoin-style 50 C3U subsidy + halvings
- proof-of-work block mining
- difficulty adjustment checkpoints
- UTXO transaction model
- Ed25519 transaction signatures
- checksummed C3U-native addresses (`c3u1…`, `tc3u1…`, `rc3u1…`)
- coinbase maturity
- transaction fees paid in C3U
- mempool validation and double-spend rejection
- JSON node API
- peer block/transaction broadcast and basic chain sync
- built-in mobile-friendly block explorer
- CLI wallet, mining, balance, transfer, and status commands
- JSON chain persistence

## Android / Termux quick start

No ETH or faucet is required.

```bash
pkg update -y
pkg install git golang -y
git clone https://github.com/Cod3Uchiha/c3u.git
cd c3u
go build -o c3u ./cmd/c3u
```

Create two **regtest** wallets:

```bash
./c3u wallet new --network regtest --out miner.wallet.json
./c3u wallet new --network regtest --out receiver.wallet.json
```

Start a local node:

```bash
./c3u node --network regtest --data ./c3udata --listen :59333
```

Open a second Termux session and mine blocks to the miner address printed above:

```bash
./c3u mine --node http://127.0.0.1:59333 --address YOUR_RC3U_ADDRESS --count 2
```

Regtest coinbase maturity is one block, so the first reward becomes spendable after the second block. Check the balance:

```bash
./c3u balance --node http://127.0.0.1:59333 --address YOUR_RC3U_ADDRESS
```

Send 1 C3U:

```bash
./c3u send \
  --node http://127.0.0.1:59333 \
  --wallet miner.wallet.json \
  --to RECEIVER_RC3U_ADDRESS \
  --amount 1 \
  --fee 0.0001
```

Mine one more block to confirm the transaction:

```bash
./c3u mine --node http://127.0.0.1:59333 --address YOUR_RC3U_ADDRESS --count 1
```

Then view the built-in explorer in your phone browser:

`http://127.0.0.1:59333/`

## Public testnet

Run:

```bash
./c3u node --network testnet --data ./c3udata --listen :49333
```

Connect nodes using repeatable peer flags:

```bash
./c3u node --network testnet --listen :49333 \
  --peer http://NODE_2_IP:49333 \
  --peer http://NODE_3_IP:49333
```

The current peer transport is HTTP/JSON for the experimental network. A hardened binary P2P protocol, peer discovery, anti-DoS controls, headers-first sync, and chain-work fork choice are required before production mainnet.

## Mainnet warning

The repository contains provisional mainnet parameters so consensus code can be tested, but **do not assign real monetary value to this implementation yet**. Before a real C3U mainnet launch, freeze the consensus specification, generate and publish final mainnet genesis constants, implement cumulative-work fork selection and hardened P2P networking, conduct external security review, and run a multi-node public testnet.

## Legacy ERC-20 prototype

The earlier Base ERC-20 experiment is preserved in the `erc20-prototype` branch. It is not the native C3U chain and is not required to run C3U Core.

## Build and test

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/c3u
```

## License

MIT.
