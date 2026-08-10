package core

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFinalMainnetGenesis(t *testing.T) {
	p, err := Params("mainnet")
	if err != nil {
		t.Fatal(err)
	}
	g := Genesis(p)
	if g.Hash != p.GenesisExpectedHash {
		t.Fatalf("mainnet genesis mismatch: got %s want %s", g.Hash, p.GenesisExpectedHash)
	}
	if !MeetsDifficulty(g.Hash, p.InitialDifficulty) {
		t.Fatal("mainnet genesis does not satisfy configured proof of work")
	}
	if p.AddressPrefix != "c3u1" || p.DefaultPort != 39333 {
		t.Fatal("unexpected mainnet identity parameters")
	}
}

func TestEncryptedWalletRoundTrip(t *testing.T) {
	w, err := NewWallet("mainnet")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "wallet.json")
	password := "correct horse battery staple"
	if err := SaveWalletEncrypted(path, w, password); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	priv, _ := base64.StdEncoding.DecodeString(w.PrivateKey)
	if bytes.Contains(b, []byte(w.PrivateKey)) || bytes.Contains(b, priv) {
		t.Fatal("wallet file exposes private key material")
	}
	unlocked, err := LoadWalletEncrypted(path, password)
	if err != nil {
		t.Fatal(err)
	}
	if unlocked.Address != w.Address || unlocked.PrivateKey != w.PrivateKey {
		t.Fatal("wallet round trip mismatch")
	}
	if _, err := LoadWalletEncrypted(path, "wrong password that is long enough"); err == nil {
		t.Fatal("wrong password unexpectedly unlocked wallet")
	}
}

func TestHigherWorkChainWins(t *testing.T) {
	p, _ := Params("regtest")
	miner, _ := NewWallet("regtest")
	genesis := Genesis(p)
	active := &Blockchain{Params: p, Blocks: []Block{genesis}}
	candidate := &Blockchain{Params: p, Blocks: []Block{genesis}}

	mineOne := func(bc *Blockchain) {
		height := bc.Tip().Height + 1
		coinbase := NewCoinbase(height, miner.Address, BlockSubsidy(height, p.HalvingInterval))
		b := Block{
			Version:    1,
			Network:    p.Name,
			Height:     height,
			PrevHash:   bc.Tip().Hash,
			MerkleRoot: MerkleRoot([]Transaction{coinbase}),
			Timestamp:  time.Now().Unix(),
			Difficulty: NextDifficulty(p, bc.Blocks),
			Txs:        []Transaction{coinbase},
		}
		MineBlock(&b)
		if err := bc.AddBlockMemory(b); err != nil {
			t.Fatal(err)
		}
	}

	mineOne(active)
	mineOne(candidate)
	mineOne(candidate)
	if ChainWork(candidate.Blocks).Cmp(ChainWork(active.Blocks)) <= 0 {
		t.Fatal("candidate should have more work")
	}
	changed, err := active.ReplaceIfBetter(candidate.Blocks)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || active.Tip().Hash != candidate.Tip().Hash {
		t.Fatal("higher-work chain was not adopted")
	}
}
