package chain_unittest

import (
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/config"
	"github.com/vitelabs/go-vite/node"
	"os"
	"path/filepath"
)

func newChainInstance(dirName string, clearDataDir bool, isStart bool) chain.Chain {
	dataDir := filepath.Join(node.DefaultDataDir(), dirName)

	if clearDataDir {
		os.RemoveAll(dataDir)
	}

	chainInstance := chain.NewChain(&config.Config{
		DataDir: dataDir,
	})
	chainInstance.Init()
	if isStart {
		chainInstance.Start()
	}

	return chainInstance
}
