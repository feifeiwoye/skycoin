package blockdb

import (
	"fmt"

	"github.com/skycoin/skycoin/src/aether/encoder"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/coin"
)

// Manages known BlockSigs as received.
// TODO -- support out of order blocks.  This requires a change to the
// message protocol to support ranges similar to bitcoin's locator hashes.
// We also need to keep track of whether a block has been executed so that
// as continuity is established we can execute chains of blocks.
// TODO -- Since we will need to hold blocks that cannot be verified
// immediately against the blockchain, we need to be able to hold multiple
// BlockSigs per BkSeq, or use hashes as keys.  For now, this is not a
// problem assuming the signed blocks created from master are valid blocks,
// because we can check the signature independently of the blockchain.
type BlockSigs struct {
	Sigs *Bucket
}

func NewBlockSigs() *BlockSigs {
	sigs, err := NewBucket([]byte("block_sigs"))
	if err != nil {
		panic(err)
	}

	return &BlockSigs{
		Sigs: sigs,
	}
}

// Checks that BlockSigs state correspond with coin.Blockchain state
// and that all signatures are valid.
func (self *BlockSigs) Verify(masterPublic cipher.PubKey, bc *coin.Blockchain) error {
	for i := uint64(0); i <= bc.Head().Seq(); i++ {
		b := bc.GetBlockInDepth(i)
		if b == nil {
			return fmt.Errorf("no block in depth %v", i)
		}
		// get sig
		sig, err := self.Get(b.HashHeader())
		if err != nil {
			return err
		}

		if err := cipher.VerifySignature(masterPublic, sig, bc.GetBlockInDepth(i).HashHeader()); err != nil {
			return err
		}
	}

	return nil
}

func (bs BlockSigs) Get(hash cipher.SHA256) (cipher.Sig, error) {
	bin := bs.Sigs.Get(hash[:])
	if bin == nil {
		return cipher.Sig{}, fmt.Errorf("no sig for %v", hash.Hex())
	}
	var sig cipher.Sig
	if err := encoder.DeserializeRaw(bin, &sig); err != nil {
		return cipher.Sig{}, err
	}
	return sig, nil
}

func (bs *BlockSigs) Add(sb *coin.SignedBlock) error {
	hash := sb.Block.HashHeader()
	return bs.Sigs.Put(hash[:], encoder.Serialize(sb.Sig))
}
