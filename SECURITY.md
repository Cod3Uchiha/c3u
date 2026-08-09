# Security

## Contract model

C3U is an ERC-20 token built with OpenZeppelin contracts. The contract is burnable, uses two-step ownership transfer, and enforces a permanent lifetime issuance ceiling of 21,000,000 C3U.

The owner can mint only from the remaining lifetime issuance allowance. `totalMinted` is cumulative: burning tokens does not increase the amount that can later be minted.

## Owner key

The owner key is security-critical because it controls the remaining unissued C3U and ownership transfer. Never commit or publish a private key, seed phrase, keystore password, or wallet backup.

For production, use a dedicated owner wallet with a secure backup. Consider transferring ownership to a multisig or hardware-backed wallet before a mainnet launch.

## Canonical contracts

The only canonical Base Sepolia test contract currently published by this repository is:

`0x47407E17bf4915b280a29AEF9e8482ae3a88Df29`

It is source-verified on BaseScan. Any other address claiming to be C3U should be treated as unrelated unless this repository explicitly lists it under `deployments/`.

There is currently **no canonical Base mainnet C3U contract**.

## Reporting

If you find a vulnerability, do not exploit it against public deployments. Open a GitHub security report/private vulnerability report when available, or contact the repository owner through their GitHub profile.

## Scope before mainnet

Before any Base mainnet deployment:

1. freeze and review the intended source commit;
2. run the complete Foundry test suite;
3. verify ownership and deployment-wallet addresses;
4. deploy and verify source on the block explorer;
5. publish the mainnet address in this repository;
6. test transfers, burns, and owner-only issuance with small amounts first.
