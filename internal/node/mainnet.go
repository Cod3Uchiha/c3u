package node

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Cod3Uchiha/c3u/internal/core"
)

const maxPeerChainResponse = 128 << 20

// SyncBestChain asks configured peers for their active chain and adopts it only
// when it is valid, has the same genesis block, and carries more cumulative PoW.
func (s *Server) SyncBestChain() {
	client := &http.Client{Timeout: 20 * time.Second}
	for _, peer := range s.Peers {
		peer = strings.TrimRight(peer, "/")
		if peer == "" {
			continue
		}

		resp, err := client.Get(peer + "/v1/status")
		if err != nil {
			continue
		}
		var status struct {
			Network string `json:"network"`
			Genesis string `json:"genesis"`
		}
		err = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&status)
		resp.Body.Close()
		if err != nil || status.Network != s.Chain.Params.Name || status.Genesis != s.Chain.Blocks[0].Hash {
			continue
		}

		resp, err = client.Get(peer + "/v1/blocks")
		if err != nil {
			continue
		}
		var blocks []core.Block
		err = json.NewDecoder(io.LimitReader(resp.Body, maxPeerChainResponse)).Decode(&blocks)
		resp.Body.Close()
		if err != nil {
			continue
		}

		s.mu.Lock()
		changed, err := s.Chain.ReplaceIfBetter(blocks)
		if changed && err == nil {
			oldPool := s.mempool
			s.mempool = map[string]core.Transaction{}
			for _, tx := range oldPool {
				_ = s.acceptTransaction(tx)
			}
		}
		s.mu.Unlock()
	}
}

// SyncLoop keeps peers converged after startup and after temporary disconnects.
func (s *Server) SyncLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	s.SyncBestChain()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.SyncBestChain()
		}
	}
}

// HardenedHandler prevents a public peer from remotely consuming the node's CPU
// through the convenience mining RPC. Public miners run their own C3U node and
// mine locally; normal block and transaction propagation remains public.
func HardenedHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		if r.ContentLength > 2<<20 {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}
		if r.URL.Path == "/v1/mine" && !remoteIsLoopback(r.RemoteAddr) {
			http.Error(w, "mining RPC is local-only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func remoteIsLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
