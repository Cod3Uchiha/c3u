package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Cod3Uchiha/c3u/internal/core"
	"github.com/Cod3Uchiha/c3u/internal/node"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "wallet":
		err = walletCmd(os.Args[2:])
	case "node":
		err = nodeCmd(os.Args[2:])
	case "mine":
		err = mineCmd(os.Args[2:])
	case "balance":
		err = balanceCmd(os.Args[2:])
	case "send":
		err = sendCmd(os.Args[2:])
	case "status":
		err = statusCmd(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`C3U Core

Commands:
  c3u wallet new --network regtest --out miner.wallet.json
  c3u node --network regtest --data ./c3udata --listen :59333
  c3u mine --node http://127.0.0.1:59333 --address rc3u1... --count 2
  c3u balance --node http://127.0.0.1:59333 --address rc3u1...
  c3u send --node http://127.0.0.1:59333 --wallet miner.wallet.json --to rc3u1... --amount 1 --fee 0.0001
  c3u status --node http://127.0.0.1:59333
`)
}

func walletCmd(args []string) error {
	if len(args) == 0 || args[0] != "new" {
		return fmt.Errorf("use: c3u wallet new")
	}
	fs := flag.NewFlagSet("wallet new", flag.ContinueOnError)
	network := fs.String("network", "regtest", "network")
	out := fs.String("out", "c3u.wallet.json", "wallet path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	w, err := core.NewWallet(*network)
	if err != nil {
		return err
	}
	if err := core.SaveWallet(*out, w); err != nil {
		return err
	}
	fmt.Println("Address:", w.Address)
	fmt.Println("Wallet:", *out)
	fmt.Println("WARNING: wallet file contains the private key. Keep it secret and back it up securely.")
	return nil
}

func nodeCmd(args []string) error {
	fs := flag.NewFlagSet("node", flag.ContinueOnError)
	network := fs.String("network", "regtest", "network")
	data := fs.String("data", "./c3udata", "data directory")
	listen := fs.String("listen", "", "listen address")
	peers := multiFlag{}
	fs.Var(&peers, "peer", "peer URL (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	params, err := core.Params(*network)
	if err != nil {
		return err
	}
	if *listen == "" {
		*listen = fmt.Sprintf(":%d", params.DefaultPort)
	}
	bc, err := core.NewBlockchain(*network, *data)
	if err != nil {
		return err
	}
	srv := node.New(bc, peers)
	go srv.Sync()
	fmt.Printf("C3U %s node listening on %s\n", *network, *listen)
	fmt.Printf("Genesis: %s\n", bc.Blocks[0].Hash)
	fmt.Printf("Explorer: http://127.0.0.1%s/\n", *listen)
	return http.ListenAndServe(*listen, srv.Handler())
}

func mineCmd(args []string) error {
	fs := flag.NewFlagSet("mine", flag.ContinueOnError)
	n := fs.String("node", "http://127.0.0.1:59333", "node URL")
	addr := fs.String("address", "", "reward address")
	count := fs.Int("count", 1, "blocks")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *addr == "" {
		return fmt.Errorf("--address required")
	}
	return postPrint(*n+"/v1/mine", map[string]any{"address": *addr, "count": *count})
}
func balanceCmd(args []string) error {
	fs := flag.NewFlagSet("balance", flag.ContinueOnError)
	n := fs.String("node", "http://127.0.0.1:59333", "node URL")
	addr := fs.String("address", "", "address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *addr == "" {
		return fmt.Errorf("--address required")
	}
	return getPrint(*n + "/v1/balance/" + *addr)
}
func statusCmd(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	n := fs.String("node", "http://127.0.0.1:59333", "node URL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return getPrint(*n + "/v1/status")
}

func sendCmd(args []string) error {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	n := fs.String("node", "http://127.0.0.1:59333", "node URL")
	walletPath := fs.String("wallet", "", "wallet file")
	to := fs.String("to", "", "recipient")
	amountS := fs.String("amount", "", "C3U amount")
	feeS := fs.String("fee", "0.0001", "fee C3U")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *walletPath == "" || *to == "" || *amountS == "" {
		return fmt.Errorf("--wallet, --to, and --amount required")
	}
	w, err := core.LoadWallet(*walletPath)
	if err != nil {
		return err
	}
	params, _ := core.Params(w.Network)
	if !core.ValidateAddress(params, *to) {
		return fmt.Errorf("recipient is not a valid %s address", w.Network)
	}
	amount, err := core.ParseAmount(*amountS)
	if err != nil {
		return err
	}
	fee, err := core.ParseAmount(*feeS)
	if err != nil {
		return err
	}
	need := amount + fee
	resp, err := http.Get(strings.TrimRight(*n, "/") + "/v1/utxos/" + w.Address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return responseError(resp)
	}
	var utxos []core.UTXO
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return err
	}
	var selected []core.UTXO
	var sum int64
	for _, u := range utxos {
		selected = append(selected, u)
		sum += u.Output.Value
		if sum >= need {
			break
		}
	}
	if sum < need {
		return fmt.Errorf("insufficient mature balance: need %s C3U", core.FormatAmount(need))
	}
	tx := core.Transaction{Timestamp: time.Now().Unix(), Outputs: []core.TxOutput{{Value: amount, Address: *to}}}
	for _, u := range selected {
		tx.Inputs = append(tx.Inputs, core.TxInput{TxID: u.TxID, Index: u.Index})
	}
	if change := sum - need; change > 0 {
		tx.Outputs = append(tx.Outputs, core.TxOutput{Value: change, Address: w.Address})
	}
	if err := tx.Sign(w); err != nil {
		return err
	}
	return postPrint(strings.TrimRight(*n, "/")+"/v1/transactions", tx)
}

func postPrint(url string, v any) error {
	b, _ := json.Marshal(v)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return responseError(resp)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}
func getPrint(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return responseError(resp)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}
func responseError(resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("node returned %s: %s", resp.Status, strings.TrimSpace(string(b)))
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}
