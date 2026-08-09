# Security status

C3U Core v0.1 is experimental blockchain software. Do not use it to secure assets with real monetary value yet.

Implemented consensus checks include block linkage, proof of work, transaction signatures, UTXO ownership, monetary-range checks, coinbase maturity, subsidy enforcement, merkle roots, and mempool double-spend rejection.

Before production mainnet, the project still needs at minimum: cumulative chain-work fork choice and reorg handling; hardened binary P2P transport and peer discovery; anti-DoS/resource limits; durable mempool behavior; wallet encryption or hardware-wallet support; richer script/multisig support; fuzz/property testing; adversarial multi-node testing; reproducible releases; and independent security review.

Never publish a wallet private key. The current CLI wallet file is intentionally simple for test networks and stores the private key locally with mode `0600`; it is not a production wallet format.
