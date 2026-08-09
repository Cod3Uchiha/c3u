package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Wallet struct {
	Network    string `json:"network"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func NewWallet(network string) (*Wallet, error) {
	params, err := Params(network)
	if err != nil {
		return nil, err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &Wallet{
		Network:    network,
		Address:    AddressFromPublicKey(params, pub),
		PublicKey:  base64.StdEncoding.EncodeToString(pub),
		PrivateKey: base64.StdEncoding.EncodeToString(priv),
	}, nil
}

func SaveWallet(path string, w *Wallet) error {
	b, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func LoadWallet(path string) (*Wallet, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var w Wallet
	if err := json.Unmarshal(b, &w); err != nil {
		return nil, err
	}
	params, err := Params(w.Network)
	if err != nil {
		return nil, err
	}
	pub, err := base64.StdEncoding.DecodeString(w.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid wallet public key")
	}
	priv, err := base64.StdEncoding.DecodeString(w.PrivateKey)
	if err != nil || len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid wallet private key")
	}
	if AddressFromPublicKey(params, pub) != w.Address {
		return nil, fmt.Errorf("wallet address does not match public key")
	}
	if !ed25519.PublicKey(pub).Equal(ed25519.PrivateKey(priv).Public()) {
		return nil, fmt.Errorf("wallet private/public key mismatch")
	}
	return &w, nil
}

func AddressFromPublicKey(params NetworkParams, pub []byte) string {
	h := sha256.Sum256(pub)
	payload := h[:20]
	chk0 := sha256.Sum256(payload)
	chk1 := sha256.Sum256(chk0[:])
	full := append(append([]byte{}, payload...), chk1[:4]...)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(full)
	return params.AddressPrefix + strings.ToLower(enc)
}

func ValidateAddress(params NetworkParams, address string) bool {
	if !strings.HasPrefix(address, params.AddressPrefix) {
		return false
	}
	raw := strings.TrimPrefix(address, params.AddressPrefix)
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(raw))
	if err != nil || len(b) != 24 {
		return false
	}
	canonical := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
	if canonical != raw {
		return false
	}
	payload := b[:20]
	chk0 := sha256.Sum256(payload)
	chk1 := sha256.Sum256(chk0[:])
	return string(b[20:]) == string(chk1[:4])
}

func DecodePrivateKey(w *Wallet) (ed25519.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(w.PrivateKey)
	if err != nil || len(b) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key")
	}
	return ed25519.PrivateKey(b), nil
}

func DecodePublicKey(s string) (ed25519.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key")
	}
	return ed25519.PublicKey(b), nil
}

func DoubleSHA256(b []byte) []byte {
	a := sha256.Sum256(b)
	c := sha256.Sum256(a[:])
	return c[:]
}

func HashHex(b []byte) string { return hex.EncodeToString(DoubleSHA256(b)) }
