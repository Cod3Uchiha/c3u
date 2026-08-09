package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Blockchain struct {
	Params NetworkParams `json:"params"`
	Blocks []Block       `json:"blocks"`
	path   string
}

func Genesis(params NetworkParams) Block {
	b := Block{
		Version:    1,
		Network:    params.Name,
		Height:     0,
		PrevHash:   "",
		MerkleRoot: MerkleRoot(nil),
		Timestamp:  params.GenesisTimestamp,
		Difficulty: params.InitialDifficulty,
		Nonce:      params.GenesisNonce,
		Extra:      "C3U 10-Aug-2026 — native money, independent network",
		Txs:        nil,
	}
	b.Hash = b.ComputeHash()
	return b
}

func NewBlockchain(network, dataDir string) (*Blockchain, error) {
	params, err := Params(network)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dataDir, network+"-chain.json")
	bc := &Blockchain{Params: params, path: path}
	if _, err := os.Stat(path); err == nil {
		if err := bc.load(); err != nil {
			return nil, err
		}
		return bc, nil
	}
	g := Genesis(params)
	if params.GenesisExpectedHash != "" && g.Hash != params.GenesisExpectedHash {
		return nil, fmt.Errorf("genesis hash mismatch: got %s", g.Hash)
	}
	if !MeetsDifficulty(g.Hash, g.Difficulty) {
		return nil, fmt.Errorf("configured genesis nonce does not satisfy proof of work")
	}
	bc.Blocks = []Block{g}
	if err := bc.Save(); err != nil {
		return nil, err
	}
	return bc, nil
}

func (bc *Blockchain) load() error {
	b, err := os.ReadFile(bc.path)
	if err != nil {
		return err
	}
	var disk struct {
		Network string  `json:"network"`
		Blocks  []Block `json:"blocks"`
	}
	if err := json.Unmarshal(b, &disk); err != nil {
		return err
	}
	if disk.Network != bc.Params.Name || len(disk.Blocks) == 0 {
		return fmt.Errorf("chain file belongs to another or invalid network")
	}
	bc.Blocks = disk.Blocks
	return bc.Verify()
}

func (bc *Blockchain) Save() error {
	payload := struct {
		Network string  `json:"network"`
		Blocks  []Block `json:"blocks"`
	}{bc.Params.Name, bc.Blocks}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	tmp := bc.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, bc.path)
}

func (bc *Blockchain) Tip() Block { return bc.Blocks[len(bc.Blocks)-1] }

func (bc *Blockchain) Verify() error {
	g := Genesis(bc.Params)
	if bc.Blocks[0].Hash != g.Hash || bc.Blocks[0].ComputeHash() != g.Hash {
		return fmt.Errorf("invalid genesis block")
	}
	utxos := map[string]UTXO{}
	for i := 1; i < len(bc.Blocks); i++ {
		if err := bc.validateBlockWithUTXO(bc.Blocks[i], bc.Blocks[:i], utxos); err != nil {
			return fmt.Errorf("block %d: %w", i, err)
		}
		applyBlockUTXO(bc.Blocks[i], utxos)
	}
	return nil
}

func (bc *Blockchain) UTXOSet() (map[string]UTXO, error) {
	utxos := map[string]UTXO{}
	for i := 1; i < len(bc.Blocks); i++ {
		if err := bc.validateBlockWithUTXO(bc.Blocks[i], bc.Blocks[:i], utxos); err != nil {
			return nil, err
		}
		applyBlockUTXO(bc.Blocks[i], utxos)
	}
	return utxos, nil
}

func (bc *Blockchain) AddBlock(block Block) error {
	utxos, err := bc.UTXOSet()
	if err != nil {
		return err
	}
	if err := bc.validateBlockWithUTXO(block, bc.Blocks, utxos); err != nil {
		return err
	}
	bc.Blocks = append(bc.Blocks, block)
	return bc.Save()
}

func (bc *Blockchain) validateBlockWithUTXO(block Block, prefix []Block, utxos map[string]UTXO) error {
	prev := prefix[len(prefix)-1]
	if block.Network != bc.Params.Name || block.Height != prev.Height+1 || block.PrevHash != prev.Hash {
		return fmt.Errorf("block does not extend current chain")
	}
	if block.Timestamp > time.Now().Unix()+2*60*60 {
		return fmt.Errorf("block timestamp too far in future")
	}
	if block.Difficulty != NextDifficulty(bc.Params, prefix) {
		return fmt.Errorf("unexpected difficulty: got %d want %d", block.Difficulty, NextDifficulty(bc.Params, prefix))
	}
	if block.ComputeHash() != block.Hash || !MeetsDifficulty(block.Hash, block.Difficulty) {
		return fmt.Errorf("invalid proof of work")
	}
	if block.MerkleRoot != MerkleRoot(block.Txs) {
		return fmt.Errorf("invalid merkle root")
	}
	if len(block.Txs) == 0 || block.Txs[0].Coinbase == "" || len(block.Txs[0].Inputs) != 0 || len(block.Txs[0].Outputs) != 1 {
		return fmt.Errorf("block requires exactly one leading coinbase transaction")
	}
	tmp := cloneUTXO(utxos)
	var fees int64
	for i := 1; i < len(block.Txs); i++ {
		tx := block.Txs[i]
		if tx.Coinbase != "" {
			return fmt.Errorf("multiple coinbase transactions")
		}
		fee, err := VerifyTransaction(bc.Params, tx, tmp, block.Height-1)
		if err != nil {
			return fmt.Errorf("tx %s: %w", tx.ID, err)
		}
		if fees > MaxMoney-fee {
			return fmt.Errorf("fee overflow")
		}
		fees += fee
		applyTxUTXO(tx, block.Height, false, tmp)
	}
	coinbase := block.Txs[0]
	if coinbase.ComputeID() != coinbase.ID {
		return fmt.Errorf("coinbase id mismatch")
	}
	out := coinbase.Outputs[0]
	if !ValidateAddress(bc.Params, out.Address) || out.Value != BlockSubsidy(block.Height, bc.Params.HalvingInterval)+fees {
		return fmt.Errorf("invalid coinbase reward")
	}
	return nil
}

func NextDifficulty(params NetworkParams, blocks []Block) uint8 {
	if len(blocks) <= 1 {
		return params.InitialDifficulty
	}
	nextHeight := blocks[len(blocks)-1].Height + 1
	if nextHeight%params.AdjustmentInterval != 0 || int64(len(blocks)) <= params.AdjustmentInterval {
		return blocks[len(blocks)-1].Difficulty
	}
	start := blocks[len(blocks)-int(params.AdjustmentInterval)]
	end := blocks[len(blocks)-1]
	actual := end.Timestamp - start.Timestamp
	expected := params.TargetBlockSeconds * params.AdjustmentInterval
	d := end.Difficulty
	if actual < expected/2 && d < 250 {
		d++
	} else if actual > expected*2 && d > 1 {
		d--
	}
	return d
}

func BlockSubsidy(height, halvingInterval int64) int64 {
	if height <= 0 || halvingInterval <= 0 {
		return 0
	}
	halvings := (height - 1) / halvingInterval
	if halvings >= 64 {
		return 0
	}
	return (50 * Coin) >> uint(halvings)
}

func IssuedSupply(height, halvingInterval int64) int64 {
	var total int64
	for h := int64(1); h <= height; h++ {
		s := BlockSubsidy(h, halvingInterval)
		if s == 0 {
			break
		}
		if total > MaxMoney-s {
			return MaxMoney
		}
		total += s
	}
	return total
}

func (bc *Blockchain) Balance(address string, matureOnly bool) (int64, error) {
	utxos, err := bc.UTXOSet()
	if err != nil {
		return 0, err
	}
	var total int64
	height := bc.Tip().Height
	for _, u := range utxos {
		if u.Output.Address != address {
			continue
		}
		if matureOnly && u.Coinbase && height-u.Height < bc.Params.CoinbaseMaturity {
			continue
		}
		total += u.Output.Value
	}
	return total, nil
}

func (bc *Blockchain) UTXOsForAddress(address string, matureOnly bool) ([]UTXO, error) {
	utxos, err := bc.UTXOSet()
	if err != nil {
		return nil, err
	}
	var out []UTXO
	height := bc.Tip().Height
	for _, u := range utxos {
		if u.Output.Address == address && (!matureOnly || !u.Coinbase || height-u.Height >= bc.Params.CoinbaseMaturity) {
			out = append(out, u)
		}
	}
	SortUTXOs(out)
	return out, nil
}

func cloneUTXO(in map[string]UTXO) map[string]UTXO {
	out := make(map[string]UTXO, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func applyTxUTXO(tx Transaction, height int64, coinbase bool, utxos map[string]UTXO) {
	for _, in := range tx.Inputs {
		delete(utxos, UTXOKey(in.TxID, in.Index))
	}
	for i, out := range tx.Outputs {
		utxos[UTXOKey(tx.ID, i)] = UTXO{TxID: tx.ID, Index: i, Output: out, Height: height, Coinbase: coinbase}
	}
}

func applyBlockUTXO(block Block, utxos map[string]UTXO) {
	for i, tx := range block.Txs {
		applyTxUTXO(tx, block.Height, i == 0, utxos)
	}
}
