package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Cod3Uchiha/c3u/internal/core"
)

type Server struct {
	Chain   *core.Blockchain
	Peers   []string
	mu      sync.Mutex
	mempool map[string]core.Transaction
}

func New(chain *core.Blockchain, peers []string) *Server {
	clean := make([]string, 0, len(peers))
	for _, p := range peers {
		p = strings.TrimRight(strings.TrimSpace(p), "/")
		if p != "" {
			clean = append(clean, p)
		}
	}
	return &Server{Chain: chain, Peers: clean, mempool: map[string]core.Transaction{}}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.explorer)
	mux.HandleFunc("/v1/status", s.status)
	mux.HandleFunc("/v1/blocks", s.blocks)
	mux.HandleFunc("/v1/block", s.receiveBlock)
	mux.HandleFunc("/v1/transactions", s.transactions)
	mux.HandleFunc("/v1/mempool", s.mempoolHandler)
	mux.HandleFunc("/v1/balance/", s.balance)
	mux.HandleFunc("/v1/utxos/", s.utxos)
	mux.HandleFunc("/v1/mine", s.mine)
	return mux
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tip := s.Chain.Tip()
	writeJSON(w, map[string]any{
		"name":             "C3U Core",
		"network":          s.Chain.Params.Name,
		"height":           tip.Height,
		"tip":              tip.Hash,
		"genesis":          s.Chain.Blocks[0].Hash,
		"difficulty_bits":  tip.Difficulty,
		"issued_satoshis":  core.IssuedSupply(tip.Height, s.Chain.Params.HalvingInterval),
		"issued_c3u":       core.FormatAmount(core.IssuedSupply(tip.Height, s.Chain.Params.HalvingInterval)),
		"mempool":          len(s.mempool),
		"peers":            s.Peers,
	})
}

func (s *Server) blocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, s.Chain.Blocks)
}

func (s *Server) receiveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var b core.Block
	if err := decodeJSON(r.Body, &b); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	err := s.Chain.AddBlock(b)
	if err == nil {
		for _, tx := range b.Txs {
			delete(s.mempool, tx.ID)
		}
	}
	s.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	go s.broadcast("/v1/block", b)
	writeJSON(w, map[string]any{"ok": true, "hash": b.Hash, "height": b.Height})
}

func (s *Server) transactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var tx core.Transaction
	if err := decodeJSON(r.Body, &tx); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	err := s.acceptTransaction(tx)
	s.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	go s.broadcast("/v1/transactions", tx)
	writeJSON(w, map[string]any{"ok": true, "txid": tx.ID})
}

func (s *Server) acceptTransaction(tx core.Transaction) error {
	if _, ok := s.mempool[tx.ID]; ok {
		return nil
	}
	utxos, err := s.Chain.UTXOSet()
	if err != nil {
		return err
	}
	spent := map[string]bool{}
	for _, m := range s.mempool {
		for _, in := range m.Inputs {
			spent[core.UTXOKey(in.TxID, in.Index)] = true
		}
	}
	for _, in := range tx.Inputs {
		if spent[core.UTXOKey(in.TxID, in.Index)] {
			return fmt.Errorf("input already spent by mempool transaction")
		}
	}
	if _, err := core.VerifyTransaction(s.Chain.Params, tx, utxos, s.Chain.Tip().Height); err != nil {
		return err
	}
	s.mempool[tx.ID] = tx
	return nil
}

func (s *Server) mempoolHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]core.Transaction, 0, len(s.mempool))
	for _, tx := range s.mempool {
		list = append(list, tx)
	}
	writeJSON(w, list)
}

func (s *Server) balance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	address := strings.TrimPrefix(r.URL.Path, "/v1/balance/")
	if !core.ValidateAddress(s.Chain.Params, address) {
		http.Error(w, "invalid address", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	confirmed, err := s.Chain.Balance(address, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	spendable, _ := s.Chain.Balance(address, true)
	writeJSON(w, map[string]any{"address": address, "confirmed_satoshis": confirmed, "confirmed_c3u": core.FormatAmount(confirmed), "spendable_satoshis": spendable, "spendable_c3u": core.FormatAmount(spendable)})
}

func (s *Server) utxos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	address := strings.TrimPrefix(r.URL.Path, "/v1/utxos/")
	if !core.ValidateAddress(s.Chain.Params, address) {
		http.Error(w, "invalid address", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.Chain.UTXOsForAddress(address, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, list)
}

type mineRequest struct {
	Address string `json:"address"`
	Count   int    `json:"count"`
}

func (s *Server) mine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req mineRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		http.Error(w, "count max is 100", 400)
		return
	}
	if !core.ValidateAddress(s.Chain.Params, req.Address) {
		http.Error(w, "invalid mining address", 400)
		return
	}
	hashes := make([]string, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		s.mu.Lock()
		b, err := s.buildAndMine(req.Address)
		if err == nil {
			err = s.Chain.AddBlock(b)
		}
		if err == nil {
			for _, tx := range b.Txs {
				delete(s.mempool, tx.ID)
			}
		}
		s.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		hashes = append(hashes, b.Hash)
		go s.broadcast("/v1/block", b)
	}
	writeJSON(w, map[string]any{"ok": true, "hashes": hashes, "height": s.Chain.Tip().Height})
}

func (s *Server) buildAndMine(address string) (core.Block, error) {
	height := s.Chain.Tip().Height + 1
	utxos, err := s.Chain.UTXOSet()
	if err != nil {
		return core.Block{}, err
	}
	tmp := make(map[string]core.UTXO, len(utxos))
	for k, v := range utxos {
		tmp[k] = v
	}
	txs := make([]core.Transaction, 0, len(s.mempool)+1)
	var fees int64
	for _, tx := range s.mempool {
		fee, err := core.VerifyTransaction(s.Chain.Params, tx, tmp, height-1)
		if err != nil {
			continue
		}
		fees += fee
		for _, in := range tx.Inputs {
			delete(tmp, core.UTXOKey(in.TxID, in.Index))
		}
		for idx, out := range tx.Outputs {
			tmp[core.UTXOKey(tx.ID, idx)] = core.UTXO{TxID: tx.ID, Index: idx, Output: out, Height: height}
		}
		txs = append(txs, tx)
	}
	coinbase := core.NewCoinbase(height, address, core.BlockSubsidy(height, s.Chain.Params.HalvingInterval)+fees)
	txs = append([]core.Transaction{coinbase}, txs...)
	b := core.Block{Version: 1, Network: s.Chain.Params.Name, Height: height, PrevHash: s.Chain.Tip().Hash, MerkleRoot: core.MerkleRoot(txs), Timestamp: time.Now().Unix(), Difficulty: core.NextDifficulty(s.Chain.Params, s.Chain.Blocks), Txs: txs}
	core.MineBlock(&b)
	return b, nil
}

func (s *Server) Sync() {
	for _, p := range s.Peers {
		resp, err := http.Get(p + "/v1/status")
		if err != nil {
			continue
		}
		var st map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		if st["network"] != s.Chain.Params.Name || st["genesis"] != s.Chain.Blocks[0].Hash {
			continue
		}
		resp, err = http.Get(p + "/v1/blocks")
		if err != nil {
			continue
		}
		var blocks []core.Block
		if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		s.mu.Lock()
		for i := len(s.Chain.Blocks); i < len(blocks); i++ {
			if err := s.Chain.AddBlock(blocks[i]); err != nil {
				break
			}
		}
		s.mu.Unlock()
	}
}

func (s *Server) broadcast(path string, v any) {
	b, _ := json.Marshal(v)
	for _, p := range s.Peers {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Post(p+path, "application/json", bytes.NewReader(b))
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func decodeJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(io.LimitReader(r, 2<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) explorer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tip := s.Chain.Tip()
	recent := s.Chain.Blocks
	if len(recent) > 12 {
		recent = recent[len(recent)-12:]
	}
	data := struct {
		Network string
		Height  int64
		Tip     string
		Supply  string
		Blocks  []core.Block
	}{s.Chain.Params.Name, tip.Height, tip.Hash, core.FormatAmount(core.IssuedSupply(tip.Height, s.Chain.Params.HalvingInterval)), recent}
	if err := explorerTemplate.Execute(w, data); err != nil {
		log.Println(err)
	}
}

var explorerTemplate = template.Must(template.New("explorer").Parse(`<!doctype html><html><head><meta name="viewport" content="width=device-width,initial-scale=1"><title>C3U Explorer</title><style>body{font-family:system-ui;background:#080b10;color:#eef2f7;margin:0;padding:24px}main{max-width:980px;margin:auto}.card{background:#101620;border:1px solid #263142;border-radius:16px;padding:18px;margin:12px 0}.muted{color:#93a4b8}.hash{font-family:monospace;word-break:break-all;font-size:12px}table{width:100%;border-collapse:collapse}td,th{padding:10px;text-align:left;border-bottom:1px solid #263142}@media(max-width:700px){.hide{display:none}}</style></head><body><main><h1>C3U Core Explorer</h1><p class="muted">Native C3U blockchain • {{.Network}}</p><div class="card"><b>Height</b> {{.Height}}<br><b>Issued</b> {{.Supply}} C3U<br><b>Tip</b><div class="hash">{{.Tip}}</div></div><div class="card"><h2>Recent blocks</h2><table><tr><th>Height</th><th>Hash</th><th class="hide">Txs</th></tr>{{range .Blocks}}<tr><td>{{.Height}}</td><td class="hash">{{.Hash}}</td><td class="hide">{{len .Txs}}</td></tr>{{end}}</table></div></main></body></html>`))
