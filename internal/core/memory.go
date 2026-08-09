package core

// AddBlockMemory validates and appends a block without disk persistence.
// It is intended for tests and embedded ephemeral chains.
func (bc *Blockchain) AddBlockMemory(block Block) error {
	utxos, err := bc.UTXOSet()
	if err != nil {
		return err
	}
	if err := bc.validateBlockWithUTXO(block, bc.Blocks, utxos); err != nil {
		return err
	}
	bc.Blocks = append(bc.Blocks, block)
	return nil
}
