# C3U Core

<p align="center"><img src="assets/c3u-logo.jpg" width="260" alt="C3U logo"></p>

**C3U Core is the native C3U blockchain.** It does not require Ethereum, Base, ERC-20, a contract address, or ETH gas. C3U is the network's native currency and transaction fees are paid in C3U.

## Mainnet status

The C3U mainnet consensus identity and genesis block are frozen for the v0.1 launch. The network becomes live when independent mainnet nodes start from this genesis and block 1 is mined.

Final mainnet genesis:

```text
000000369992bbd8b1c7df0c1298529357c4e5a564b3355afbd6c7f2d2ee67b4
```

C3U Core is new software and has not had the years of public review of mature cryptocurrency clients. A one-node chain is technically a mainnet but is not decentralized or strongly attack-resistant. Security improves as independent nodes and miners join.

## Monetary policy

- Native currency: **C3U**
- Smallest unit: **0.00000001 C3U**
- Precision: **8 decimals**
- Initial block subsidy: **50 C3U**
- Halving interval: **210,000 blocks**
- Target mainnet block interval: **10 minutes**
- Mainnet difficulty retarget: **144 blocks**
- Coinbase maturity: **100 blocks**
- Maximum monetary range: **21,000,000 C3U**
- Genesis premine: **none**
- Consensus: **Proof of Work**
- Ledger: **UTXO**

The subsidy schedule approaches but never exceeds 21 million C3U because rewards are calculated in integer 1e-8 C3U units.

## Mainnet security changes

The launch candidate adds:

- encrypted wallet private keys at rest using AES-256-GCM
- PBKDF2-HMAC-SHA256 password-based key derivation
- cumulative-proof-of-work fork choice
- continuous peer reconciliation instead of startup-only sync
- fixed mainnet genesis constants
- higher initial mainnet PoW difficulty
- local-only mining RPC to prevent public peers from remotely consuming node CPU
- HTTP server timeouts and request-size limits

## Build on Android / Termux

```bash
pkg update -y
pkg install git golang -y
git clone https://github.com/Cod3Uchiha/c3u.git
cd c3u
go build -o c3u ./cmd/c3u
```

## Create a real mainnet wallet

Do not place the password directly in the command line. Read it silently into an environment variable:

```bash
read -s -p "C3U wallet password: " C3U_WALLET_PASSWORD; echo
export C3U_WALLET_PASSWORD
./c3u wallet new --network mainnet --out c3u-main.wallet.json
unset C3U_WALLET_PASSWORD
```

A mainnet address begins with `c3u1`.

Back up `c3u-main.wallet.json` and the password separately. The private key is encrypted in the wallet file; losing the password can make the coins unrecoverable.

## Start C3U Mainnet

```bash
./c3u node \
  --network mainnet \
  --data ./c3u-mainnet \
  --listen :39333
```

Local explorer:

```text
http://127.0.0.1:39333/
```

Check status:

```bash
./c3u status --node http://127.0.0.1:39333
```

## Mine native C3U

```bash
./c3u mine \
  --node http://127.0.0.1:39333 \
  --address YOUR_C3U1_ADDRESS \
  --count 1
```

Block 1 creates the first **50 native C3U**. There is no premine. Mainnet mining rewards become spendable after 100 blocks.

## Connect independent nodes

```bash
./c3u node \
  --network mainnet \
  --data ./c3u-mainnet \
  --listen :39333 \
  --peer http://NODE_2_IP:39333 \
  --peer http://NODE_3_IP:39333
```

Configured peers repeatedly compare chains. A replacement chain must have the same C3U mainnet genesis block, pass full consensus validation, and carry strictly more cumulative proof of work.

## Send C3U

```bash
read -s -p "C3U wallet password: " C3U_WALLET_PASSWORD; echo
export C3U_WALLET_PASSWORD
./c3u send \
  --node http://127.0.0.1:39333 \
  --wallet c3u-main.wallet.json \
  --to RECIPIENT_C3U1_ADDRESS \
  --amount 1 \
  --fee 0.0001
unset C3U_WALLET_PASSWORD
```

## Networks

| Network | Address prefix | Port | Purpose |
| --- | --- | ---: | --- |
| Mainnet | `c3u1` | 39333 | Production C3U chain |
| Testnet | `tc3u1` | 49333 | Public testing |
| Regtest | `rc3u1` | 59333 | Local development |

Full frozen mainnet specification and operating instructions are in [`docs/MAINNET.md`](docs/MAINNET.md).

The old Base ERC-20 experiment remains preserved on the `erc20-prototype` branch and is not part of the native C3U network.

## Build and test

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/c3u
```

## License

MIT.
