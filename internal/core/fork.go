package core

import (
	"fmt"
	"math/big"
)

// ChainWork returns the cumulative proof-of-work represented by a chain.
// With C3U's leading-zero-bit difficulty model, one block at difficulty n
// contributes 2^n units of work.
func ChainWork(blocks []Block) *big.Int {
	total := new(big.Int)
	one := big.NewInt(1)
	for _, b := range blocks {
		work := new(big.Int).Lsh(new(big.Int).Set(one), uint(b.Difficulty))
		total.Add(total, work)
	}
	return total
}

func ChainWorkString(blocks []Block) string {
	return ChainWork(blocks).String()
}

// ReplaceIfBetter atomically adopts a fully validated candidate chain only
// when it carries strictly more cumulative proof-of-work than the active chain.
func (bc *Blockchain) ReplaceIfBetter(blocks []Block) (bool, error) {
	if len(blocks) == 0 {
		return false, fmt.Errorf("candidate chain is empty")
	}
	if len(bc.Blocks) == 0 {
		return false, fmt.Errorf("active chain is empty")
	}
	if blocks[0].Hash != bc.Blocks[0].Hash {
		return false, fmt.Errorf("candidate chain has a different genesis block")
	}

	candidateBlocks := append([]Block(nil), blocks...)
	candidate := &Blockchain{Params: bc.Params, Blocks: candidateBlocks}
	if err := candidate.Verify(); err != nil {
		return false, fmt.Errorf("candidate chain invalid: %w", err)
	}
	if ChainWork(candidateBlocks).Cmp(ChainWork(bc.Blocks)) <= 0 {
		return false, nil
	}

	old := bc.Blocks
	bc.Blocks = candidateBlocks
	if bc.path != "" {
		if err := bc.Save(); err != nil {
			bc.Blocks = old
			return false, err
		}
	}
	return true, nil
}
