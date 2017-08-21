// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"sync"

	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type BlockMaker struct {
	Mutex sync.RWMutex

	NumberOfLeaders int

	ProcessedEBEntries  []*EBlockEntry
	ProcessedFBEntries  []interfaces.ITransaction
	ProcessedABEntries  []interfaces.IABEntry
	ProcessedECBEntries []*ECBlockEntry

	VMs map[int]*VM

	BState *blockchainState.BlockchainState

	ABlockHeaderExpansionArea []byte

	CurrentMinute int
}

func NewBlockMaker() *BlockMaker {
	bm := new(BlockMaker)
	bm.NumberOfLeaders = 1
	bm.BState = blockchainState.NewBSLocalNet()
	return bm
}

func (bm *BlockMaker) SetCurrentMinute(m int) {
	bm.CurrentMinute = m
}

type MsgAckPair struct {
	Message interfaces.IMessageWithEntry
	Ack     *messages.Ack
}

func ChainIDToVMIndex(h interfaces.IHash, numberOfLeaders int) int {
	hash := h.Bytes()

	if numberOfLeaders < 2 {
		return 0
	}

	v := uint64(0)
	for _, b := range hash {
		v += uint64(b)
	}

	r := int(v % uint64(numberOfLeaders))
	return r
}

type VM struct {
	Mutex sync.RWMutex

	DBHeight uint32

	LatestHeight uint32
	LatestAck    *messages.Ack

	PendingPairs []*MsgAckPair
}

func (bm *BlockMaker) GetVM(chainID interfaces.IHash) *VM {
	bm.Mutex.Lock()
	defer bm.Mutex.Unlock()

	index := ChainIDToVMIndex(chainID, bm.NumberOfLeaders)
	vm := bm.VMs[index]
	if vm == nil {
		vm = new(VMs)
		bm.VMs[index] = vm
	}

	return vm
}

func (bm *BlockMaker) ProcessAckedMessage(msg interfaces.IMessageWithEntry, ack *messages.Ack) error {
	chainID := msg.GetEntryChainID()
	vm := bm.GetVM(chainID)

	vm.Mutex.Lock()
	defer vm.Mutex.Unlock()

	if ack.Height < vm.LatestHeight {
		//We already processed this message, nothing to do
		return nil
	}
	if ack.Height == vm.LatestHeight {
		if vm.LatestAck != nil {
			//We already processed this message as well
			//AND it's not the first message!
			//Nothing to do
			return nil
		}
	}

	//Insert message into the slice, then process off of slice one by one
	//This is to reduce complexity of the code
	pair := new(MsgAckPair)
	pair.Ack = ack
	pair.Message = msg

	inserted := false
	for i := 0; i < len(vm.PendingPairs); i++ {
		//Looking for first pair that is higher than the current Height, so we can insert our pair before the other one
		if vm.PendingPairs[i].Ack.Height > pair.Ack.Height {
			index := i - 1
			if index < 0 {
				//Inserting as the first entry
				vm.PendingPairs = append([]*MsgAckPair{pair}, vm.PendingPairs...)
			} else {
				//Inserting somewhere in the middle
				vm.PendingPairs = append(vm.PendingPairs[:index], append([]*MsgAckPair{pair}, vm.PendingPairs[index:]...))
			}
			break
		}
		if vm.PendingPairs[i].Ack.Height == pair.Ack.Height {
			//TODO: figure out what to do when an ACK has the same height
			//If it's not the same or something?
		}
	}
	if inserted == false {
		vm.PendingPairs = append(vm.PendingPairs, pair)
	}

	//Iterate over pending pairs and process them one by one until we're stuck
	for {
		if len(vm.PendingPairs) == 0 {
			break
		}
		if vm.LatestAck == nil {
			if vm.PendingPairs[0].Ack.Height != 0 {
				//We're expecting first message and we didn't find one
				break
			}
		} else {
			if vm.LatestHeight != vm.PendingPairs[0].Ack.Height-1 {
				//We didn't find the next pair
				break
			}
		}

		pair = vm.PendingPairs[0]
		ok, err := pair.Ack.VerifySerialHash(vm.LatestAck)
		if err != nil {
			return err
		}
		if ok == false {
			//TODO: reject the ACK or something?
		}

		//Actually processing the message
		//TODO: do
		switch chainID.String() {
		case "000000000000000000000000000000000000000000000000000000000000000a":
			break
		case "000000000000000000000000000000000000000000000000000000000000000c":
			break
		case "000000000000000000000000000000000000000000000000000000000000000f":
			break
		default:
			break
		}
	}

	return nil
}