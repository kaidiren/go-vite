package vm

import (
	"bytes"
	"encoding/hex"
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
	"testing"
)

func TestRun(t *testing.T) {
	vm := &VM{StateDb: &NoDatabase{}, createBlock: CreateNoVmAccountBlock, instructionSet: simpleInstructionSet}
	vm.Debug = true
	// return 1+2
	inputdata, _ := hex.DecodeString("6001600201602080919052602090F3")
	receiveCallBlock := CreateNoVmAccountBlock(types.Address{}, types.Address{}, BlockTypeReceive, 1)
	c := newContract(receiveCallBlock.AccountAddress(), receiveCallBlock.ToAddress(), receiveCallBlock, 1000000, 0)
	c.setCallCode(types.Address{}, inputdata)
	ret, _ := c.run(vm)
	expectedRet, _ := hex.DecodeString("03")
	expectedRet = leftPadBytes(expectedRet, 32)
	if bytes.Compare(ret, expectedRet) != 0 || c.quotaLeft != 999964 || c.quotaRefund != 0 {
		t.Fatalf("expected [%v], get [%v]", expectedRet, ret)
	}
}

func TestVM_CreateSend(t *testing.T) {
	inputdata, _ := hex.DecodeString("608060405260008055348015601357600080fd5b5060358060216000396000f3006080604052600080fd00a165627a7a723058207c31c74808fe0f95820eb3c48eac8e3e10ef27058dc6ca159b547fccde9290790029")
	sendCreateBlock := CreateNoVmAccountBlock(types.Address{}, types.Address{}, BlockTypeSendCreate, 1)
	sendCreateBlock.SetTokenId(viteTokenTypeId)
	sendCreateBlock.SetAmount(big.NewInt(10))
	sendCreateBlock.SetSnapshotHash(types.Hash{})
	sendCreateBlock.SetPrevHash(types.Hash{})
	sendCreateBlock.SetHeight(big.NewInt(1))
	sendCreateBlock.SetData(inputdata)
	sendCreateBlock.SetCreateFee(big.NewInt(0))
	// vm.Debug = true
	vm := NewVM(&NoDatabase{}, CreateNoVmAccountBlock, VMConfig{})
	blockList, _, err := vm.Run(sendCreateBlock)
	if len(blockList) != 1 ||
		//blockList[0].Quota() != 58336 ||
		blockList[0].ToAddress() == emptyAddress ||
		//blockList[0].Balance() == nil ||
		blockList[0].Amount().Cmp(big.NewInt(10)) != 0 ||
		//blockList[0].StateHash() == emptyHash ||
		blockList[0].TokenId() != viteTokenTypeId {
		t.Fatalf("send create fail [%v] %v", blockList, err)
	}
}

/*func TestVM_CreateReceive(t *testing.T) {
	inputdata, _ := hex.DecodeString("608060405260008055348015601357600080fd5b5060358060216000396000f3006080604052600080fd00a165627a7a723058207c31c74808fe0f95820eb3c48eac8e3e10ef27058dc6ca159b547fccde9290790029")
	receiveCreateBlock := CreateNoVmAccountBlock(types.Address{}, types.Address{}, BlockTypeReceive, 1)
	receiveCreateBlock.SetTokenId(viteTokenTypeId)
	receiveCreateBlock.SetAmount(big.NewInt(0))
	receiveCreateBlock.SetSnapshotHash(types.Hash{})
	receiveCreateBlock.SetPrevHash(types.Hash{})
	receiveCreateBlock.SetHeight(big.NewInt(1))
	receiveCreateBlock.SetData(inputdata)
	vm := NewVM(&NoDatabase{}, CreateNoVmAccountBlock)
	vm.Debug = true
	blockList, err := vm.Run(receiveCreateBlock)
	if len(blockList) != 1 ||
		//blockList[0].Quota() != 89008 ||
		//blockList[0].Balance() == nil ||
		//blockList[0].StateHash() == emptyHash ||
		blockList[0].Amount().Cmp(big.NewInt(0)) != 0 ||
		blockList[0].TokenId() != viteTokenTypeId {
		t.Fatalf("send create fail [%v] %v", blockList, err)
	}
}*/
