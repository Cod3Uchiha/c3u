package core

import (
	"testing"
	"time"
)

func TestAmountRoundTrip(t *testing.T) {
	v, err := ParseAmount("123.45678901")
	if err != nil {
		t.Fatal(err)
	}
	if v != 12_345_678_901 {
		t.Fatalf("got %d", v)
	}
	if FormatAmount(v) != "123.45678901" {
		t.Fatal(FormatAmount(v))
	}
}
func TestAddressChecksum(t *testing.T) {
	w, err := NewWallet("regtest")
	if err != nil {
		t.Fatal(err)
	}
	p, _ := Params("regtest")
	if !ValidateAddress(p, w.Address) {
		t.Fatal("address should validate")
	}
	bad := w.Address[:len(w.Address)-1] + "a"
	if bad == w.Address {
		bad = w.Address[:len(w.Address)-1] + "b"
	}
	if ValidateAddress(p, bad) {
		t.Fatal("bad checksum accepted")
	}
}
func TestBitcoinStyleIssuanceStaysUnder21M(t *testing.T) {
	p, _ := Params("mainnet")
	var total int64
	for h := int64(1); ; h++ {
		s := BlockSubsidy(h, p.HalvingInterval)
		if s == 0 {
			break
		}
		total += s
		if total > MaxMoney {
			t.Fatalf("exceeded max at %d", h)
		}
	}
	if total <= 20_999_000*Coin {
		t.Fatalf("unexpectedly low issuance %s", FormatAmount(total))
	}
}
func TestRegtestMineAndSpend(t *testing.T) {
	miner, _ := NewWallet("regtest")
	receiver, _ := NewWallet("regtest")
	p, _ := Params("regtest")
	g := Genesis(p)
	bc := &Blockchain{Params: p, Blocks: []Block{g}}
	mine := func(addr string) {
		h := bc.Tip().Height + 1
		cb := NewCoinbase(h, addr, BlockSubsidy(h, p.HalvingInterval))
		b := Block{Version: 1, Network: p.Name, Height: h, PrevHash: bc.Tip().Hash, MerkleRoot: MerkleRoot([]Transaction{cb}), Timestamp: time.Now().Unix(), Difficulty: NextDifficulty(p, bc.Blocks), Txs: []Transaction{cb}}
		MineBlock(&b)
		if err := bc.AddBlockMemory(b); err != nil {
			t.Fatal(err)
		}
	}
	mine(miner.Address)
	mine(miner.Address)
	utxos, err := bc.UTXOsForAddress(miner.Address, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(utxos) == 0 {
		t.Fatal("expected mature utxo")
	}
	u := utxos[0]
	tx := Transaction{Timestamp: time.Now().Unix(), Inputs: []TxInput{{TxID: u.TxID, Index: u.Index}}, Outputs: []TxOutput{{Value: Coin, Address: receiver.Address}, {Value: u.Output.Value - Coin - 10_000, Address: miner.Address}}}
	if err := tx.Sign(miner); err != nil {
		t.Fatal(err)
	}
	set, _ := bc.UTXOSet()
	if _, err := VerifyTransaction(p, tx, set, bc.Tip().Height); err != nil {
		t.Fatal(err)
	}
}
