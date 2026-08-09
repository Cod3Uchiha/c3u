# C3U Core protocol draft v0.1

## Ledger
C3U uses a UTXO ledger. A transaction consumes existing outputs and creates new outputs. Inputs prove control of the referenced output by presenting an Ed25519 public key whose C3U address matches the output and a valid signature over the unsigned transaction payload.

## Units
`1 C3U = 100,000,000` atomic units. Consensus amounts are signed 64-bit integers and must remain in the `[0, 21,000,000 C3U]` monetary range.

## Issuance
Block 0 is a zero-premine genesis block. Height 1 begins a 50 C3U subsidy. On mainnet/testnet the subsidy halves every 210,000 blocks. Subsidy uses integer atomic units and eventually reaches zero. Consensus never creates a block subsidy above the schedule.

## Proof of work
Block hashes are double-SHA-256. A block is valid when its hash has at least the configured number of leading zero bits. Difficulty is represented as leading-zero bits in v0.1 and is adjusted at network-specific intervals by at most one bit per adjustment.

This is deliberately simpler than Bitcoin's compact-target `nBits` representation. Mainnet launch should freeze a target representation and cumulative-chain-work rule before public value is attached.

## Addresses
Mainnet addresses begin `c3u1`, testnet `tc3u1`, and regtest `rc3u1`. The payload is the first 20 bytes of SHA-256(public key) plus a four-byte double-SHA-256 checksum, Base32 encoded without padding.

## Coinbase maturity
Mainnet: 100 blocks. Testnet: 20 blocks. Regtest: 1 block.

## Networks
- mainnet: port 39333, 600-second target, initial 18-bit PoW
- testnet: port 49333, 60-second target, initial 12-bit PoW
- regtest: port 59333, 1-second target, initial 4-bit PoW

Mainnet parameters are provisional until the public testnet is hardened and the final genesis constants are frozen.
