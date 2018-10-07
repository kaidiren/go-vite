package model

import (
	"container/list"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/contracts"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/vm_context"
	"sync"
	"time"
)

const (
	fullCacheExpireTime   = 2 * time.Minute
	simpleCacheExpireTime = 20 * time.Minute
)

// obtaining the account info from cache or db and manage the cache lifecycle
type OnroadBlocksPool struct {
	dbAccess *UAccess

	fullCache          *sync.Map // map[types.Address]*onroadBlocksCache
	fullCacheDeadTimer *sync.Map // map[types.Address]*time.Timer

	simpleCache          *sync.Map // map[types.Address]*CommonAccountInfo
	simpleCacheDeadTimer *sync.Map //map[types.Address]*time.Timer

	newCommonTxListener   map[types.Address]func()
	commonTxListenerMutex sync.RWMutex

	newContractListener   map[types.Gid]func()
	contractListenerMutex sync.RWMutex

	log log15.Logger
}

func NewOnroadBlocksPool(dbAccess *UAccess) *OnroadBlocksPool {
	return &OnroadBlocksPool{
		dbAccess:             dbAccess,
		fullCache:            &sync.Map{},
		fullCacheDeadTimer:   &sync.Map{},
		simpleCache:          &sync.Map{},
		simpleCacheDeadTimer: &sync.Map{},
		newCommonTxListener:  make(map[types.Address]func()),
		newContractListener:  make(map[types.Gid]func()),
		log:                  log15.New("onroad", "OnroadBlocksPool"),
	}
}

func (p *OnroadBlocksPool) GetAddrListByGid(gid types.Gid) (addrList []*types.Address, err error) {
	return p.dbAccess.GetContractAddrListByGid(&gid)
}

func (p *OnroadBlocksPool) Close() error {
	p.log.Info("Close()")

	p.simpleCacheDeadTimer.Range(func(_, value interface{}) bool {
		if value != nil {
			value.(*time.Timer).Stop()
		}
		return true
	})
	p.simpleCache = nil

	p.fullCacheDeadTimer.Range(func(_, value interface{}) bool {
		if value != nil {
			value.(*time.Timer).Stop()
		}
		return true
	})
	p.fullCache = nil

	p.log.Info("Close() end")
	return nil
}

func (p *OnroadBlocksPool) addSimpleCache(addr types.Address, accountInfo *CommonAccountInfo) {
	//p.log.Info("addSimpleCache", "addr", addr, "TotalNumber", accountInfo.TotalNumber)
	p.simpleCache.Store(addr, accountInfo)

	timer, ok := p.simpleCacheDeadTimer.Load(addr)
	if ok && timer != nil {
		p.log.Info("addSimpleCache Reset timer")
		timer.(*time.Timer).Reset(simpleCacheExpireTime)
	} else {
		p.simpleCacheDeadTimer.Store(addr, time.AfterFunc(simpleCacheExpireTime, func() {
			p.log.Info("simple cache end life delete it", "addr", addr)
			p.simpleCache.Delete(addr)
		}))
	}
}

func (p *OnroadBlocksPool) GetCommonAccountInfo(addr types.Address) (*CommonAccountInfo, error) {
	p.log.Info("first load in simple cache", "addr", addr)
	if c, ok := p.simpleCache.Load(addr); ok {
		v, ok := p.simpleCacheDeadTimer.Load(addr)
		if ok {
			v.(*time.Timer).Reset(simpleCacheExpireTime)
		}
		return c.(*CommonAccountInfo), nil
	}

	p.log.Info("second load from full cache", "addr", addr)
	// fixme: getTokenId func
	//if fullcache, ok := p.fullCache.Load(addr); ok {
	//	accountInfo := fullcache.(*onroadBlocksCache).toCommonAccountInfo(p.dbAccess.Chain.GetTokenInfoById)
	//	if accountInfo != nil {
	//		p.addSimpleCache(addr, accountInfo)
	//		return accountInfo, nil
	//	}
	//}

	p.log.Info("third load from db", "addr", addr)
	accountInfo, e := p.dbAccess.GetCommonAccInfo(&addr)
	if e != nil {
		return nil, e
	}
	if accountInfo != nil {
		p.addSimpleCache(addr, accountInfo)
	}

	return accountInfo, nil

}

func (p *OnroadBlocksPool) GetNextTx(addr types.Address) *ledger.AccountBlock {
	p.log.Info("GetNextTx", "addr", addr)
	c, ok := p.fullCache.Load(addr)
	if !ok {
		return nil
	}
	return c.(*onroadBlocksCache).GetNextTx()
}

func (p *OnroadBlocksPool) ResetCacheCursor(addr types.Address) {
	p.log.Info("ResetCacheCursor", "addr", addr)
	c, ok := p.fullCache.Load(addr)
	if !ok {
		return
	}
	c.(*onroadBlocksCache).ResetCursor()
}

func (p *OnroadBlocksPool) AcquireAccountInfoCache(addr types.Address) error {
	log := p.log.New("AcquireAccountInfoCache", addr)
	if t, ok := p.fullCacheDeadTimer.Load(addr); ok {
		if t != nil {
			log.Info("stop timer")
			t.(*time.Timer).Stop()
		}
	}

	if c, ok := p.fullCache.Load(addr); ok {
		c.(*onroadBlocksCache).addReferenceCount()
		log.Info("found in cache", "ref", c.(*onroadBlocksCache).referenceCount)
		return nil
	}

	blocks, e := p.dbAccess.GetAllOnroadBlocks(addr)
	if e != nil {
		log.Error("get from db", "err", e)
		return e
	}
	log.Info("get from db", "len", len(blocks))

	list := list.New()
	for _, value := range blocks {
		list.PushBack(value)
	}

	p.fullCache.Store(addr, &onroadBlocksCache{
		blocks:         *list,
		currentEle:     list.Front(),
		referenceCount: 1,
	})

	return nil
}

func (p *OnroadBlocksPool) ReleaseAccountInfoCache(addr types.Address) error {
	log := p.log.New("ReleaseAccountInfoCache", addr)
	v, ok := p.fullCache.Load(addr)
	if !ok {
		log.Info("no cache found")
		return nil
	}
	c := v.(*onroadBlocksCache)
	if c.subReferenceCount() <= 0 {
		log.Info("cache found ref <= 0 delete cache")

		c.ResetCursor()
		p.fullCacheDeadTimer.Store(addr, time.AfterFunc(fullCacheExpireTime, func() {
			log.Info("cache delete")
			p.DeleteFullCache(addr)
		}))
		return nil
	}
	log.Info("after release", "ref", c.referenceCount)

	return nil
}

func (p *OnroadBlocksPool) DeleteFullCache(address types.Address) {
	p.fullCache.Delete(address)
}

func (p *OnroadBlocksPool) WriteOnroad(batch *leveldb.Batch, blockList []*vm_context.VmAccountBlock) error {
	p.log.Info("WriteOnroad ")

	for _, v := range blockList {
		if v.AccountBlock.IsSendBlock() {
			// basic writeMeta func
			if err := p.dbAccess.writeOnroadMeta(batch, v.AccountBlock); err != nil {
				p.log.Error("writeOnroadMeta", "error", err)
				return err
			}

			if v.AccountBlock.BlockType == ledger.BlockTypeSendCreate {
				unsavedCache := v.VmContext.UnsavedCache()
				gidList := unsavedCache.ContractGidList()
				for _, v := range gidList {
					if err := p.dbAccess.WriteContractAddrToGid(batch, *v.Gid(), *v.Addr()); err != nil {
						p.log.Error("WriteContractAddrToGid", "error", err)
						return err
					}
				}
			}
		} else {
			if err := p.dbAccess.deleteOnroadMeta(batch, v.AccountBlock); err != nil {
				p.log.Error("deleteOnroadMeta", "error", err)
				return err
			}
		}
	}
	// todo 确认写好之后 再更新
	// p.updateCache(writeType, block)
	return nil
}

func (p *OnroadBlocksPool) DeleteDirect(sendBlock *ledger.AccountBlock) error {
	return p.dbAccess.store.DeleteMeta(nil, &sendBlock.ToAddress, &sendBlock.Hash)
}

type signBlock struct {
	block  *ledger.AccountBlock
	ignore bool
}

func copyToSignLedger(subLedger map[types.Address][]*ledger.AccountBlock) map[types.Address][]*signBlock {
	var handlerLedger map[types.Address][]*signBlock
	for k, blockList := range subLedger {
		for _, v := range blockList {
			handlerLedger[k] = append(handlerLedger[k], &signBlock{v, false})
		}
	}
	return handlerLedger
}

// DeleteUnRoad means to revert according to bifurcation
func (p *OnroadBlocksPool) DeleteOnroad(batch *leveldb.Batch, subLedger map[types.Address][]*ledger.AccountBlock) error {
	p.log.Info("DeleteOnroad: revert")

	signLedger := copyToSignLedger(subLedger)
	for _, blockList := range signLedger {
		// the blockList is sorted by height with ascending order
		for i := len(blockList); i > 0; i-- {
			v := blockList[i]

			if v.block.IsReceiveBlock() {
				if v.ignore == true {
					continue
				}

				sendBlock, err := p.dbAccess.Chain.GetAccountBlockByHash(&v.block.FromBlockHash)
				if err != nil {
					p.log.Error("GetAccountBlockByHash", "error", err)
					return err
				}
				if err := p.dbAccess.writeOnroadMeta(batch, sendBlock); err != nil {
					p.log.Error("revert receiveBlock failed", "error", err)
					return err
				}
			} else {
				existInSubLedger(v.block, signLedger)
				if err := p.dbAccess.deleteOnroadMeta(batch, v.block); err != nil {
					p.log.Error("revert the sendBlock's and the referred failed", "error", err)
					return err
				}

				// fixme: wait for func @yd
				// delete the gid-contractAddrList relationship
				gid := contracts.GetGidFromCreateContractData(v.block.Data)
				p.dbAccess.DeleteContractAddrFromGid(batch, gid, v.block.ToAddress)
			}
		}
	}
	return nil
}

func existInSubLedger(block *ledger.AccountBlock, ignoreLedger map[types.Address][]*signBlock) {
	if block.IsSendBlock() {
		if bl, ok := ignoreLedger[block.ToAddress]; ok {
			for _, v := range bl {
				if v.block.FromBlockHash == block.Hash {
					v.ignore = true
				}
			}
		}
	}
}

func (p *OnroadBlocksPool) updateFullCache(writeType bool, block *ledger.AccountBlock) error {
	v, ok := p.fullCache.Load(block.ToAddress)
	fullCache := v.(*onroadBlocksCache)
	// todo check == 0
	if !ok || fullCache.blocks.Len() == 0 {
		//p.log.Info("updateCache：no fullCache")
		return nil
	}

	if writeType {
		fullCache.addTx(block)
	} else {
		fullCache.rmTx(block)
	}

	return nil
}

// todo add mutex
func (p *OnroadBlocksPool) updateSimpleCache(writeType bool, block *ledger.AccountBlock) error {

	value, ok := p.simpleCache.Load(block.ToAddress)
	if !ok {
		// p.log.Info("updateSimpleCache：no cache")
		return nil
	}
	simpleAccountInfo := value.(*CommonAccountInfo)

	tokenBalanceInfo, ok := simpleAccountInfo.TokenBalanceInfoMap[block.TokenId]
	if writeType {
		if ok {
			tokenBalanceInfo.TotalAmount.Add(&tokenBalanceInfo.TotalAmount, block.Amount)
			tokenBalanceInfo.Number += 1
		} else {
			// fixme: remove token info?
			//token, err := p.dbAccess.Chain.GetTokenInfoById(&block.TokenId)
			//if err != nil {
			//	return errors.New("func UpdateCommonAccInfo.GetByTokenId failed" + err.Error())
			//}
			//if token == nil {
			//	return errors.New("func UpdateCommonAccInfo.GetByTokenId failed token nil")
			//}
			//simpleAccountInfo.TokenBalanceInfoMap[block.TokenId].Token = *token
			//simpleAccountInfo.TokenBalanceInfoMap[block.TokenId].TotalAmount = *block.Amount
			//simpleAccountInfo.TokenBalanceInfoMap[block.TokenId].Number = 1
		}
		simpleAccountInfo.TotalNumber += 1
	} else {
		if ok {
			if tokenBalanceInfo.TotalAmount.Cmp(block.Amount) == -1 {
				return errors.New("conflict with the memory info, so can't update when writeType is false")
			}
			if tokenBalanceInfo.TotalAmount.Cmp(block.Amount) == 0 {
				delete(simpleAccountInfo.TokenBalanceInfoMap, block.TokenId)
			} else {
				tokenBalanceInfo.TotalAmount.Sub(&tokenBalanceInfo.TotalAmount, block.Amount)
			}
		} else {
			p.log.Info("find no memory tokenInfo, so can't update when writeType is false")
		}
		simpleAccountInfo.TotalNumber -= 1
		tokenBalanceInfo.Number -= 1
	}

	return nil
}

func (p *OnroadBlocksPool) updateCache(writeType bool, block *ledger.AccountBlock) {
	e := p.updateFullCache(writeType, block)
	if e != nil {
		p.log.Error("updateFullCache", "err", e)
	}

	e = p.updateSimpleCache(writeType, block)
	if e != nil {
		p.log.Error("updateSimpleCache", "err", e)
	}
}

func (p *OnroadBlocksPool) NewSignalToWorker(block *ledger.AccountBlock) {
	gid, err := p.dbAccess.Chain.GetContractGid(&block.AccountAddress)
	if err != nil {
		p.log.Error("NewSignalToWorker", "err", err)
		return
	}
	if gid != nil {
		p.contractListenerMutex.RLock()
		defer p.contractListenerMutex.RUnlock()
		if f, ok := p.newContractListener[*gid]; ok {
			f()
		}
	} else {
		p.commonTxListenerMutex.RLock()
		defer p.commonTxListenerMutex.RUnlock()
		if f, ok := p.newCommonTxListener[block.ToAddress]; ok {
			f()
		}
	}
}

func (p *OnroadBlocksPool) GetOnroadBlocks(index, num, count uint64, addr *types.Address) (blockList []*ledger.AccountBlock, err error) {
	return p.dbAccess.GetOnroadBlocks(index, num, count, addr)
}

func (p *OnroadBlocksPool) AddCommonTxLis(addr types.Address, f func()) {
	p.commonTxListenerMutex.Lock()
	defer p.commonTxListenerMutex.Unlock()
	p.newCommonTxListener[addr] = f
}

func (p *OnroadBlocksPool) RemoveCommonTxLis(addr types.Address) {
	p.commonTxListenerMutex.Lock()
	defer p.commonTxListenerMutex.Unlock()
	delete(p.newCommonTxListener, addr)
}

func (p *OnroadBlocksPool) AddContractLis(gid types.Gid, f func()) {
	p.contractListenerMutex.Lock()
	defer p.contractListenerMutex.Unlock()
	p.newContractListener[gid] = f
}

func (p *OnroadBlocksPool) RemoveContractLis(gid types.Gid) {
	p.contractListenerMutex.Lock()
	defer p.contractListenerMutex.Unlock()
	delete(p.newContractListener, gid)
}