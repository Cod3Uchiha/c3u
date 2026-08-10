package core

import "fmt"

type NetworkParams struct {
	Name                string `json:"name"`
	AddressPrefix       string `json:"address_prefix"`
	DefaultPort         int    `json:"default_port"`
	InitialDifficulty   uint8  `json:"initial_difficulty_bits"`
	TargetBlockSeconds  int64  `json:"target_block_seconds"`
	AdjustmentInterval  int64  `json:"adjustment_interval"`
	HalvingInterval     int64  `json:"halving_interval"`
	CoinbaseMaturity    int64  `json:"coinbase_maturity"`
	GenesisTimestamp    int64  `json:"genesis_timestamp"`
	GenesisNonce        uint64 `json:"genesis_nonce"`
	GenesisExpectedHash string `json:"genesis_expected_hash"`
}

func Params(name string) (NetworkParams, error) {
	switch name {
	case "mainnet":
		return NetworkParams{
			Name:                "mainnet",
			AddressPrefix:       "c3u1",
			DefaultPort:         39333,
			InitialDifficulty:   26,
			TargetBlockSeconds:  600,
			AdjustmentInterval:  144,
			HalvingInterval:     210000,
			CoinbaseMaturity:    100,
			GenesisTimestamp:    1786337460,
			GenesisNonce:        7140716,
			GenesisExpectedHash: "000000369992bbd8b1c7df0c1298529357c4e5a564b3355afbd6c7f2d2ee67b4",
		}, nil
	case "testnet":
		return NetworkParams{Name: "testnet", AddressPrefix: "tc3u1", DefaultPort: 49333, InitialDifficulty: 12, TargetBlockSeconds: 60, AdjustmentInterval: 144, HalvingInterval: 210000, CoinbaseMaturity: 20, GenesisTimestamp: 1786314661, GenesisNonce: 9194, GenesisExpectedHash: "0003db54495abfc6ea3aeaa445c7ed1ebb1dac521853ebefa39061f328e55aad"}, nil
	case "regtest":
		return NetworkParams{Name: "regtest", AddressPrefix: "rc3u1", DefaultPort: 59333, InitialDifficulty: 4, TargetBlockSeconds: 1, AdjustmentInterval: 144, HalvingInterval: 150, CoinbaseMaturity: 1, GenesisTimestamp: 1786314662, GenesisNonce: 3, GenesisExpectedHash: "06167c2a285c6564dbf0254788e311f181f772dd3d9c3532cbde077531fe83ae"}, nil
	default:
		return NetworkParams{}, fmt.Errorf("unknown network %q (use mainnet, testnet, or regtest)", name)
	}
}
