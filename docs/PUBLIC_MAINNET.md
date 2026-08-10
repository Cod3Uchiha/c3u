# C3U Public Mainnet Deployment

C3U is not deployed like a web application. A public launch means multiple always-on C3U nodes are reachable from the internet, miners can connect to them, users can download verified node binaries, and the network converges on the valid chain with the greatest cumulative proof of work.

## Before block 1

Do not mine mainnet block 1 until the public bootstrap infrastructure is online. Once the chain begins, changing consensus parameters or the genesis block creates an incompatible network.

Recommended minimum launch topology:

- 3 independent public bootstrap nodes
- separate hosts or failure domains where possible
- static public IPv4/IPv6 addresses
- TCP port 39333 reachable from the internet
- one public explorer endpoint
- HTTPS reverse proxy for explorer/API access if desired
- public release binaries and SHA-256 checksums

The current C3U peer transport uses HTTP/JSON over TCP. Peer addresses are supplied explicitly with repeated `--peer` flags. DNS-based automatic seed discovery is not yet implemented, so publish the bootstrap peer URLs in the release notes and documentation.

## Suggested DNS names

Point DNS A/AAAA records at the public nodes, for example:

- `seed1.example.com`
- `seed2.example.com`
- `seed3.example.com`
- `explorer.example.com`

Do not put the node behind a serverless platform. The C3U node must remain running and retain its chain data.

## Install a bootstrap node on Linux

Create a dedicated user and data directory:

```bash
sudo useradd --system --home /var/lib/c3u --shell /usr/sbin/nologin c3u
sudo mkdir -p /var/lib/c3u
sudo chown c3u:c3u /var/lib/c3u
```

Install the C3U binary at `/usr/local/bin/c3u`, then install `deploy/c3u.service`:

```bash
sudo cp c3u /usr/local/bin/c3u
sudo chmod 0755 /usr/local/bin/c3u
sudo cp deploy/c3u.service /etc/systemd/system/c3u.service
sudo systemctl daemon-reload
sudo systemctl enable c3u
```

For bootstrap nodes, edit the service with peer arguments so each node knows the others. Example Node 1 command:

```bash
/usr/local/bin/c3u node \
  --network mainnet \
  --data /var/lib/c3u \
  --listen :39333 \
  --peer http://seed2.example.com:39333 \
  --peer http://seed3.example.com:39333
```

Node 2 and Node 3 should point back to the other bootstrap nodes.

Open the C3U network port in the host firewall:

```bash
sudo ufw allow 39333/tcp
```

Start the service and verify it:

```bash
sudo systemctl start c3u
sudo systemctl status c3u
curl http://127.0.0.1:39333/v1/status
```

All bootstrap nodes must report the same mainnet genesis hash:

```text
000000369992bbd8b1c7df0c1298529357c4e5a564b3355afbd6c7f2d2ee67b4
```

## Docker

Build:

```bash
docker build -t c3u-core .
```

Run:

```bash
docker run -d \
  --name c3u \
  --restart unless-stopped \
  -p 39333:39333 \
  -v c3u-mainnet:/data \
  c3u-core
```

To add peers, override the container command and append repeated `--peer` arguments.

## Publish release binaries

The repository contains `.github/workflows/release.yml`. Push a signed version tag only after CI is green:

```bash
git tag -s v0.1.0 -m "C3U Core v0.1.0"
git push origin v0.1.0
```

GitHub Actions builds Linux amd64/arm64, Windows amd64, and macOS amd64/arm64 binaries and publishes SHA-256 checksum files with the release.

## Start the chain

Only after all bootstrap nodes are reachable and agreeing on genesis:

1. Create a mainnet mining wallet and back up the encrypted wallet file.
2. Run the miner against a bootstrap node you control.
3. Mine block 1 to the mainnet `c3u1...` mining address.
4. Confirm the new block propagates to every bootstrap node.
5. Publish the block 1 hash and release information.
6. Invite independent operators to run nodes and miners using the published peers.

The first block subsidy is 50 C3U. Coinbase outputs require 100-block maturity on mainnet before they can be spent.

## Operational rules

- Never publish wallet files, passwords, private keys, or seed material.
- Back up public-node chain data, but keep mining-wallet backups separate from node servers.
- Monitor disk, RAM, CPU, uptime, peer reachability, height, tip hash, and cumulative chain work.
- Keep bootstrap nodes on identical C3U Core versions during launch.
- Do not silently change consensus constants after launch; consensus changes require an explicit network upgrade.
- Treat the current network as an early mainnet implementation until it has had sustained multi-node operation and independent security review.

## What remains before Bitcoin-level networking

C3U has proof-of-work chain validation and cumulative-work fork choice, but its peer layer is intentionally simpler than Bitcoin Core. A later networking milestone should add native binary P2P framing, automatic peer discovery/DNS seeds, peer address gossip, connection limits, anti-DoS scoring, headers-first synchronization, and stronger eclipse/Sybil resistance.
