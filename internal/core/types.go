package core

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TxInput struct {
	TxID      string `json:"txid"`
	Index     int    `json:"index"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

type TxOutput struct {
	Value   int64  `json:"value"`
	Address string `json:"address"`
}

type Transaction struct {
	ID        string     `json:"id"`
	Timestamp int64      `json:"timestamp"`
	Coinbase  string     `json:"coinbase,omitempty"`
	Inputs    []TxInput  `json:"inputs,omitempty"`
	Outputs   []TxOutput `json:"outputs"`
}

type Block struct {
	Version    uint32        `json:"version"`
	Network    string        `json:"network"`
	Height     int64         `json:"height"`
	PrevHash   string        `json:"prev_hash"`
	MerkleRoot string        `json:"merkle_root"`
	Timestamp  int64         `json:"timestamp"`
	Difficulty uint8         `json:"difficulty_bits"`
	Nonce      uint64        `json:"nonce"`
	Extra      string        `json:"extra,omitempty"`
	Txs        []Transaction `json:"transactions"`
	Hash       string        `json:"hash"`
}

type UTXO struct {
	TxID     string   `json:"txid"`
	Index    int      `json:"index"`
	Output   TxOutput `json:"output"`
	Height   int64    `json:"height"`
	Coinbase bool     `json:"coinbase"`
}

func NewCoinbase(height int64, address string, value int64) Transaction {
	tx := Transaction{
		Timestamp: time.Now().Unix(),
		Coinbase:  fmt.Sprintf("C3U block %d", height),
		Outputs:   []TxOutput{{Value: value, Address: address}},
	}
	tx.ID = tx.ComputeID()
	return tx
}

func (tx Transaction) signingBytes() []byte {
	copyTx := tx
	copyTx.ID = ""
	copyTx.Inputs = append([]TxInput(nil), tx.Inputs...)
	for i := range copyTx.Inputs {
		copyTx.Inputs[i].Signature = ""
	}
	b, _ := json.Marshal(copyTx)
	return b
}

func (tx Transaction) ComputeID() string {
	copyTx := tx
	copyTx.ID = ""
	b, _ := json.Marshal(copyTx)
	return HashHex(b)
}

func (tx *Transaction) Sign(w *Wallet) error {
	priv, err := DecodePrivateKey(w)
	if err != nil {
		return err
	}
	pub := priv.Public().(ed25519.PublicKey)
	pub64 := base64.StdEncoding.EncodeToString(pub)
	for i := range tx.Inputs {
		tx.Inputs[i].PublicKey = pub64
		tx.Inputs[i].Signature = ""
	}
	msg := tx.signingBytes()
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, msg))
	}
	tx.ID = tx.ComputeID()
	return nil
}

func VerifyTransaction(params NetworkParams, tx Transaction, utxos map[string]UTXO, currentHeight int64) (int64, error) {
	if tx.Coinbase != "" {
		return 0, fmt.Errorf("coinbase must be validated as part of a block")
	}
	if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
		return 0, fmt.Errorf("transaction requires inputs and outputs")
	}
	if tx.ComputeID() != tx.ID {
		return 0, fmt.Errorf("transaction id mismatch")
	}
	msg := tx.signingBytes()
	seen := map[string]bool{}
	var inputSum, outputSum int64
	for _, in := range tx.Inputs {
		key := UTXOKey(in.TxID, in.Index)
		if seen[key] {
			return 0, fmt.Errorf("duplicate input %s", key)
		}
		seen[key] = true
		u, ok := utxos[key]
		if !ok {
			return 0, fmt.Errorf("missing input %s", key)
		}
		if u.Coinbase && currentHeight-u.Height < params.CoinbaseMaturity {
			return 0, fmt.Errorf("coinbase input %s is immature", key)
		}
		pub, err := DecodePublicKey(in.PublicKey)
		if err != nil {
			return 0, err
		}
		if AddressFromPublicKey(params, pub) != u.Output.Address {
			return 0, fmt.Errorf("input public key does not own %s", key)
		}
		sig, err := base64.StdEncoding.DecodeString(in.Signature)
		if err != nil || len(sig) != ed25519.SignatureSize || !ed25519.Verify(pub, msg, sig) {
			return 0, fmt.Errorf("invalid signature for %s", key)
		}
		if u.Output.Value <= 0 || inputSum > MaxMoney-u.Output.Value {
			return 0, fmt.Errorf("invalid input value")
		}
		inputSum += u.Output.Value
	}
	for _, out := range tx.Outputs {
		if out.Value <= 0 || !ValidateAddress(params, out.Address) {
			return 0, fmt.Errorf("invalid transaction output")
		}
		if outputSum > MaxMoney-out.Value {
			return 0, fmt.Errorf("output value overflow")
		}
		outputSum += out.Value
	}
	if outputSum > inputSum {
		return 0, fmt.Errorf("outputs exceed inputs")
	}
	return inputSum - outputSum, nil
}

func UTXOKey(txid string, index int) string { return txid + ":" + strconv.Itoa(index) }

func MerkleRoot(txs []Transaction) string {
	if len(txs) == 0 {
		return HashHex(nil)
	}
	level := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		b, err := hex.DecodeString(tx.ID)
		if err != nil {
			b = DoubleSHA256([]byte(tx.ID))
		}
		level = append(level, b)
	}
	for len(level) > 1 {
		if len(level)%2 == 1 {
			level = append(level, append([]byte{}, level[len(level)-1]...))
		}
		next := make([][]byte, 0, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			pair := append(append([]byte{}, level[i]...), level[i+1]...)
			next = append(next, DoubleSHA256(pair))
		}
		level = next
	}
	return hex.EncodeToString(level[0])
}

func (b Block) headerBytes() []byte {
	parts := []string{
		strconv.FormatUint(uint64(b.Version), 10), b.Network, strconv.FormatInt(b.Height, 10), b.PrevHash,
		b.MerkleRoot, strconv.FormatInt(b.Timestamp, 10), strconv.Itoa(int(b.Difficulty)), strconv.FormatUint(b.Nonce, 10), b.Extra,
	}
	return []byte(strings.Join(parts, "|"))
}

func (b Block) ComputeHash() string { return HashHex(b.headerBytes()) }

func MeetsDifficulty(hash string, bits uint8) bool {
	b, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	remaining := bits
	for _, v := range b {
		if remaining == 0 {
			return true
		}
		if remaining >= 8 {
			if v != 0 {
				return false
			}
			remaining -= 8
			continue
		}
		mask := byte(0xff << (8 - remaining))
		return v&mask == 0
	}
	return remaining == 0
}

func MineBlock(b *Block) {
	for {
		b.Hash = b.ComputeHash()
		if MeetsDifficulty(b.Hash, b.Difficulty) {
			return
		}
		b.Nonce++
	}
}

func SortUTXOs(v []UTXO) {
	sort.Slice(v, func(i, j int) bool {
		if v[i].Height == v[j].Height {
			if v[i].TxID == v[j].TxID {
				return v[i].Index < v[j].Index
			}
			return v[i].TxID < v[j].TxID
		}
		return v[i].Height < v[j].Height
	})
}
