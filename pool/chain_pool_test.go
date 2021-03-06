package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/pool/tree"
)

//
//import (
//	"testing"
//
//	"github.com/vitelabs/go-vite/log15"
//)
//
//type mockChainPool struct {
//	c *forkedChain
//}
//
//func (self *mockChainPool) insertBlock(block commonBlock) error {
//	self.c.addHead(block)
//	return nil
//}
//
//func (self *mockChainPool) insertBlocks(blocks []commonBlock) error {
//	for _, v := range blocks {
//		self.c.addHead(v)
//	}
//	return nil
//}
//
//func (self *mockChainPool) head() commonBlock {
//	return self.c.Head()
//}
//
//func (self *mockChainPool) getBlock(height uint64) commonBlock {
//	return self.c.getBlock(height, false)
//}
//
//func TestChainPool(t *testing.T) {
//	mock := &mockChainPool{c: mockChain(nil, 1, 1, 10)}
//	mock.c.referChain = nil
//
//	diskChain := &diskChain{chainId: "diskchain", rw: mock, v: &ForkVersion{}}
//	cp := &chainPool{
//		poolID:    "chain-Pool-Id",
//		diskChain: diskChain,
//		log:       log15.New("module", "mock"),
//	}
//	cp.current = &forkedChain{}
//	cp.current.chainId = cp.genChainID()
//	cp.init()
//
//	printChain(cp.current)
//
//	cp.current.addHead(newMockCommonBlock(1, 10))
//
//	printChain(cp.current)
//	cp.writeToChain(cp.current, cp.current.GetBlock(10))
//
//	printChain(cp.current)
//
//	cp.current.addHead(newMockCommonBlock(1, 11))
//	cp.current.addHead(newMockCommonBlock(1, 12))
//	cp.current.addHead(newMockCommonBlock(1, 13))
//
//	c := mockChain(cp.current, 2, 12, 18)
//	cp.addChain(c)
//
//	cp.currentModifyToChain(c)
//
//	printChain(c)
//
//	bc := BCPool{chainpool: cp}
//
//	var rms []commonBlock
//	rms = append(rms, mock.c.getBlock(10, false))
//	rms = append(rms, mock.c.getBlock(9, false))
//	for _, v := range rms {
//		mock.c.removeHead(v)
//	}
//	bc.rollbackCurrent(rms)
//
//	printChain(cp.current)
//
//	printChain(mock.c)
//
//}
//
//func TestBCPool_CurrentModifyToChain(t *testing.T) {
//	mock := &mockChainPool{c: mockChain(nil, 1, 1, 10)}
//	mock.c.referChain = nil
//
//	diskChain := &diskChain{chainId: "diskchain", rw: mock, v: &ForkVersion{}}
//	cp := &chainPool{
//		poolID:    "chain-Pool-Id",
//		diskChain: diskChain,
//		log:       log15.New("module", "mock"),
//	}
//	cp.current = &forkedChain{}
//	cp.current.chainId = cp.genChainID()
//	cp.init()
//
//	tmps := mockBlocks(1, 10, 20)
//	for _, v := range tmps {
//		cp.current.addHead(v)
//	}
//
//	c2 := mockChain(cp.current, 2, 11, 25)
//	cp.addChain(c2)
//
//	c3 := mockChain(c2, 3, 13, 29)
//	cp.addChain(c3)
//
//	c4 := mockChain(c3, 4, 15, 29)
//	cp.addChain(c4)
//
//	cp.writeToChain(cp.current, cp.current.GetBlock(10))
//	cp.writeToChain(cp.current, cp.current.GetBlock(11))
//	cp.writeToChain(cp.current, cp.current.GetBlock(12))
//	cp.writeToChain(cp.current, cp.current.GetBlock(13))
//	cp.writeToChain(cp.current, cp.current.GetBlock(14))
//
//	reduceChainByRefer(c4)
//	err := cp.currentModifyToChain(c4)
//	if err != nil {
//		t.Error(err)
//	}
//
//	printChain(cp.current)
//
//	cp.check()
//}
//
//func TestChainPoolModifyRefer(t *testing.T) {
//	mock := &mockChainPool{c: mockChain(nil, 1, 1, 10)}
//	mock.c.referChain = nil
//
//	diskChain := &diskChain{chainId: "diskchain", rw: mock, v: &ForkVersion{}}
//	cp := &chainPool{
//		poolID:    "chain-Pool-Id",
//		diskChain: diskChain,
//		log:       log15.New("module", "mock"),
//	}
//	cp.current = &forkedChain{}
//	cp.current.chainId = "c1"
//	cp.init()
//
//	tmps := mockChain(mock.c, 1, 11, 20)
//	for i := tmps.tailHeight + 1; i <= tmps.headHeight; i++ {
//		cp.current.addHead(tmps.getHeightBlock(i))
//	}
//
//	c2 := mockChain(cp.current, 2, 13, 25)
//	c2.chainId = "c2"
//	cp.addChain(c2)
//
//	c3 := mockChain(c2, 3, 15, 29)
//	c3.chainId = "c3"
//	cp.addChain(c3)
//
//	c4 := mockChain(c3, 4, 16, 32)
//	c4.chainId = "c4"
//	cp.addChain(c4)
//
//	//printChain(cp.current)
//	//printChain(c2)
//	printChainJust(c3)
//	printChainJust(c4)
//
//	//cp.modifyRefer(c3, c4)
//	cp.currentModifyToChain(c4)
//
//	printChainJust(c3)
//	printChainJust(c4)
//
//	//println(c3.referChain.id(), c3.id())
//	//println(c4.referChain.id(), c4.id())
//
//	cp.check()
//}
//
//func TestChainPoolModifyRefer2(t *testing.T) {
//	mock := &mockChainPool{c: mockChain(nil, 1, 1, 10)}
//	mock.c.referChain = nil
//
//	diskChain := &diskChain{chainId: "diskchain", rw: mock, v: &ForkVersion{}}
//	cp := &chainPool{
//		poolID:    "chain-Pool-Id",
//		diskChain: diskChain,
//		log:       log15.New("module", "mock"),
//	}
//	cp.current = &forkedChain{}
//	cp.current.chainId = "c1"
//	cp.init()
//	tmps := mockChain(mock.c, 1, 11, 15)
//	for i := tmps.tailHeight + 1; i <= tmps.headHeight; i++ {
//		cp.current.addHead(tmps.getHeightBlock(i))
//	}
//
//	c2 := mockChain(cp.current, 2, 11, 25)
//	c2.chainId = "c2"
//	cp.addChain(c2)
//
//	c3 := mockChain(mock.c, 3, 9, 10)
//	c3.chainId = "c3"
//	c3.referChain = cp.current
//	cp.addChain(c3)
//
//	//printChain(cp.current)
//	//printChain(c2)
//	printChainJust(c3)
//	printChainJust(c2)
//	printChainJust(cp.current)
//
//	cp.modifyRefer(cp.current, c2)
//
//	printChainJust(c3)
//	printChainJust(c2)
//	printChainJust(cp.current)
//
//	//println(c3.referChain.id(), c3.id())
//	//println(c4.referChain.id(), c4.id())
//	//cp.current = c2
//	//cp.check()
//	//cp.modifyChainRefer()
//	//cp.check()
//}
//func TestChainPoolModifyRefer3(t *testing.T) {
//	mock := &mockChainPool{c: mockChain(nil, 1, 1, 10)}
//	mock.c.referChain = nil
//
//	diskChain := &diskChain{chainId: "diskchain", rw: mock, v: &ForkVersion{}}
//	cp := &chainPool{
//		poolID:    "chain-Pool-Id",
//		diskChain: diskChain,
//		log:       log15.New("module", "mock"),
//	}
//	cp.current = &forkedChain{}
//	cp.current.chainId = "c1"
//	cp.init()
//	tmps := mockChain(mock.c, 1, 11, 15)
//	for i := tmps.tailHeight + 1; i <= tmps.headHeight; i++ {
//		cp.current.addHead(tmps.getHeightBlock(i))
//	}
//
//	c2 := mockChain(cp.current, 2, 11, 25)
//	c2.chainId = "c2"
//	cp.addChain(c2)
//
//	c3 := mockChain(mock.c, 3, 9, 10)
//	c3.chainId = "c3"
//	c3.referChain = cp.current
//	cp.addChain(c3)
//
//	//printChain(cp.current)
//	//printChain(c2)
//	printChainJust(c3)
//	printChainJust(c2)
//	printChainJust(cp.current)
//
//	cp.modifyRefer(cp.current, c2)
//
//	printChainJust(c3)
//	printChainJust(c2)
//	printChainJust(cp.current)
//
//	//println(c3.referChain.id(), c3.id())
//	//println(c4.referChain.id(), c4.id())
//	//cp.current = c2
//	//cp.check()
//	//cp.modifyChainRefer()
//	//cp.check()
//}

func Test_Forkable(t *testing.T) {
	tr := tree.NewTree()
	diskChain := tree.NewMockBranchRoot()
	{
		// init root
		flag := "root"
		for i := 0; i < 5; i++ {
			height, hash := diskChain.HeadHH()
			diskChain.AddHead(newMockCommonBlockByHH(height, hash, flag))
		}
		height, _ := diskChain.HeadHH()
		assert.Equal(t, height, uint64(5))
	}

	cp := &chainPool{
		poolID: "unittest",
		//diskChain: diskChain,
		tree: tr,
		log:  log15.New("module", "unittest"),
	}
	cp.snippetChains = make(map[string]*snippetChain)
	cp.tree.Init(cp.poolID, diskChain)

	main := tr.Main()
	height, _ := main.HeadHH()
	assert.Equal(t, height, uint64(5))
	height, _ = main.TailHH()
	assert.Equal(t, height, uint64(5))

	for i := 0; i < 6; i++ {
		height, hash := main.HeadHH()
		err := tr.AddHead(main, newMockCommonBlockByHH(height, hash, "main"))
		assert.Empty(t, err)
	}

	height, _ = main.HeadHH()
	assert.Equal(t, height, uint64(11))
	height, _ = main.TailHH()
	assert.Equal(t, height, uint64(5))

	knot := main.GetKnot(5, true)

	block := newMockCommonBlockByHH(knot.Height(), knot.Hash(), "snippet")
	snp := newSnippetChain(block, "snippet1")

	bs := tr.Branches()
	bm := make(map[string]tree.Branch)
	for _, v := range bs {
		bm[v.ID()] = v
	}
	forky, insertable, c, err := cp.fork2(snp, bm, nil)

	assert.Empty(t, err)
	assert.NotEmpty(t, c)
	assert.Equal(t, c.ID(), main.ID())
	assert.False(t, insertable)
	assert.True(t, forky)
}

func Test_Forkable2(t *testing.T) {
	tr := tree.NewTree()
	diskChain := tree.NewMockBranchRoot()
	{
		height, _ := diskChain.HeadHH()
		assert.Equal(t, height, uint64(0))
	}

	cp := &chainPool{
		poolID: "unittest",
		//diskChain: diskChain,
		tree: tr,
		log:  log15.New("module", "unittest"),
	}
	cp.snippetChains = make(map[string]*snippetChain)
	cp.tree.Init(cp.poolID, diskChain)

	main := tr.Main()
	height, _ := main.HeadHH()
	assert.Equal(t, height, uint64(0))
	height, _ = main.TailHH()
	assert.Equal(t, height, uint64(0))

	block := newMockCommonBlockByHH(0, types.Hash{}, "snippet")
	snp := newSnippetChain(block, "snippet1")

	bs := tr.Branches()
	bm := make(map[string]tree.Branch)
	for _, v := range bs {
		bm[v.ID()] = v
	}
	forky, insertable, c, err := cp.fork2(snp, bm, nil)

	assert.Empty(t, err)
	assert.NotEmpty(t, c)
	assert.Equal(t, c.ID(), main.ID())
	assert.True(t, insertable)
	assert.False(t, forky)
}
