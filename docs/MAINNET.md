# C3U Mainnet

C3U Mainnet is the production network for the native C3U blockchain. It is independent of Ethereum, Base, ERC-20, and any smart-contract platform.

## Frozen network identity

- Network: `mainnet`
- Native currency: `C3U`
- Address prefix: `c3u1`
- Default node port: `39333`
- Precision: 8 decimals
- Smallest unit: 0.00000001 C3U
- Proof of work: double-SHA256 header hashing with leading-zero-bit difficulty
- Target block interval: 600 seconds
- Initial difficulty: 26 bits
- Difficulty adjustment interval: 144 blocks
- Initial subsidy: 50 C3U
- Halving interval: 210,000 blocks
- Coinbase maturity: 100 blocks
- Maximum monetary range: 21,000,000 C3U
- Premine: none

## Final genesis block

- Timestamp: `1786337460`
- Nonce: `7140716`
- Hash: `000000369992bbd8b1c7df0c1298529357c4e5a564b3355afbd6c7f2d2ee67b4`
- Message: `C3U 10-Aug-2026 — native money, independent network`

Changing any consensus parameter or the genesis constants creates an incompatible network. Mainnet nodes must reject a chain with a different genesis hash.

## Wallet security

Mainnet wallets created by the CLI are encrypted at rest with AES-256-GCM. A password-derived key uses PBKDF2-HMAC-SHA256 with a random salt. The password is never stored in the wallet file.

Set the wallet password without putting it directly in a command argument:

```bash
read -s -p "C3U wallet password: " C3U_WALLET_PASSWORD; echo
export C3U_WALLET_PASSWORD
./c3u wallet new --network mainnet --out c3u-main.wallet.json
unset C3U_WALLET_PASSWORD
```

Back up the encrypted wallet file and password separately. Losing the password means losing access to the private key.

## Start a mainnet node

```bash
./c3u node \
  --network mainnet \
  --data ./c3u-mainnet \
  --listen :39333
```

The built-in explorer is available locally at:

```text
http://127.0.0.1:39333/
```

For a multi-node network, configure known peers explicitly:

```bash
./c3u node \
  --network mainnet \
  --data ./c3u-mainnet \
  --listen :39333 \
  --peer http://NODE_2_IP:39333 \
  --peer http://NODE_3_IP:39333
```

Configured peers are checked repeatedly. A candidate chain is adopted only if it validates against C3U consensus, has the same mainnet genesis block, and contains strictly more cumulative proof of work.

## Mine mainnet C3U

Mining through the convenience `/v1/mine` endpoint is restricted to the local machine so a public peer cannot remotely consume the node's CPU.

```bash
./c3u mine \
  --node http://127.0.0.1:39333 \
  --address YOUR_C3U1_ADDRESS \
  --count 1
```

Block 1 creates the first 50 C3U. Mainnet coinbase outputs require 100 blocks of maturity before they can be spent.

## Send C3U

Unlock an encrypted wallet for the duration of the send command using the environment variable:

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

## Network-security note

A mainnet can technically begin with one node/miner, but decentralization and resistance to chain attacks come from independent nodes and independent proof-of-work. C3U Core is new software and has not received the years of review and adversarial testing of mature cryptocurrency implementations. Do not represent C3U as having Bitcoin-level security merely because its monetary schedule is Bitcoin-like.
