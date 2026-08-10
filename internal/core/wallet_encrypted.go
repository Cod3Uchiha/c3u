package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
)

const (
	walletFileVersion   = 2
	walletKDFIterations = 600_000
)

type encryptedWalletFile struct {
	Version       int    `json:"version"`
	Network       string `json:"network"`
	Address       string `json:"address"`
	PublicKey     string `json:"public_key"`
	Cipher        string `json:"cipher"`
	KDF           string `json:"kdf"`
	KDFIterations int    `json:"kdf_iterations"`
	Salt          string `json:"salt"`
	Nonce         string `json:"nonce"`
	Ciphertext    string `json:"ciphertext"`
}

// SaveWalletEncrypted stores only encrypted private-key material on disk.
// The address and public key remain visible so wallets can be identified
// without unlocking them.
func SaveWalletEncrypted(path string, w *Wallet, password string) error {
	if len(password) < 12 {
		return fmt.Errorf("wallet password must be at least 12 characters")
	}
	if err := validateWalletMaterial(w); err != nil {
		return err
	}
	priv, err := base64.StdEncoding.DecodeString(w.PrivateKey)
	if err != nil || len(priv) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid wallet private key")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	key := pbkdf2SHA256([]byte(password), salt, walletKDFIterations, 32)
	defer zeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	aad := walletAAD(w.Network, w.Address, w.PublicKey)
	ciphertext := gcm.Seal(nil, nonce, priv, aad)
	zeroBytes(priv)

	disk := encryptedWalletFile{
		Version:       walletFileVersion,
		Network:       w.Network,
		Address:       w.Address,
		PublicKey:     w.PublicKey,
		Cipher:        "AES-256-GCM",
		KDF:           "PBKDF2-HMAC-SHA256",
		KDFIterations: walletKDFIterations,
		Salt:          base64.StdEncoding.EncodeToString(salt),
		Nonce:         base64.StdEncoding.EncodeToString(nonce),
		Ciphertext:    base64.StdEncoding.EncodeToString(ciphertext),
	}
	b, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Chmod(path, 0o600)
}

// LoadWalletEncrypted unlocks a version-2 encrypted wallet. Legacy plaintext
// wallets remain readable on regtest/testnet so existing development wallets
// continue to work, but plaintext mainnet wallets are explicitly rejected.
func LoadWalletEncrypted(path, password string) (*Wallet, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var probe struct {
		Version    int    `json:"version"`
		Network    string `json:"network"`
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return nil, err
	}
	if probe.PrivateKey != "" {
		if probe.Network == "mainnet" {
			return nil, fmt.Errorf("refusing plaintext mainnet wallet; create an encrypted wallet")
		}
		return LoadWallet(path)
	}

	var disk encryptedWalletFile
	if err := json.Unmarshal(b, &disk); err != nil {
		return nil, err
	}
	if disk.Version != walletFileVersion || disk.Cipher != "AES-256-GCM" || disk.KDF != "PBKDF2-HMAC-SHA256" {
		return nil, fmt.Errorf("unsupported wallet format")
	}
	if disk.KDFIterations < 100_000 || disk.KDFIterations > 2_000_000 {
		return nil, fmt.Errorf("invalid wallet KDF parameters")
	}
	if password == "" {
		return nil, fmt.Errorf("wallet is encrypted; set C3U_WALLET_PASSWORD")
	}

	salt, err := base64.StdEncoding.DecodeString(disk.Salt)
	if err != nil || len(salt) != 16 {
		return nil, fmt.Errorf("invalid wallet salt")
	}
	nonce, err := base64.StdEncoding.DecodeString(disk.Nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet nonce")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(disk.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet ciphertext")
	}

	key := pbkdf2SHA256([]byte(password), salt, disk.KDFIterations, 32)
	defer zeroBytes(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid wallet nonce size")
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, walletAAD(disk.Network, disk.Address, disk.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("unable to unlock wallet: wrong password or corrupted wallet")
	}
	defer zeroBytes(plain)
	if len(plain) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid decrypted private key")
	}
	w := &Wallet{
		Network:    disk.Network,
		Address:    disk.Address,
		PublicKey:  disk.PublicKey,
		PrivateKey: base64.StdEncoding.EncodeToString(plain),
	}
	if err := validateWalletMaterial(w); err != nil {
		return nil, err
	}
	return w, nil
}

func validateWalletMaterial(w *Wallet) error {
	params, err := Params(w.Network)
	if err != nil {
		return err
	}
	pub, err := base64.StdEncoding.DecodeString(w.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid wallet public key")
	}
	priv, err := base64.StdEncoding.DecodeString(w.PrivateKey)
	if err != nil || len(priv) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid wallet private key")
	}
	if AddressFromPublicKey(params, pub) != w.Address {
		return fmt.Errorf("wallet address does not match public key")
	}
	if !ed25519.PublicKey(pub).Equal(ed25519.PrivateKey(priv).Public()) {
		return fmt.Errorf("wallet private/public key mismatch")
	}
	return nil
}

func walletAAD(network, address, publicKey string) []byte {
	return []byte(network + "|" + address + "|" + publicKey)
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	const hashLen = 32
	blocks := (keyLen + hashLen - 1) / hashLen
	out := make([]byte, 0, blocks*hashLen)
	for blockIndex := 1; blockIndex <= blocks; blockIndex++ {
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(salt)
		var counter [4]byte
		binary.BigEndian.PutUint32(counter[:], uint32(blockIndex))
		_, _ = mac.Write(counter[:])
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
		zeroBytes(u)
		zeroBytes(t)
	}
	return out[:keyLen]
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
