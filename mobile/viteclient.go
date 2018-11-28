package mobile

import (
	"context"
	"encoding/json"
	"github.com/vitelabs/go-vite/rpc"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

type AccountInfo struct {
	api.RpcAccountInfo
}

type Client struct {
	c *rpc.Client
}

func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

func NewClient(c *rpc.Client) *Client {
	return &Client{c}
}

func (vc *Client) Close() {
	vc.c.Close()
}

func (vc *Client) GetBlocksByAccAddr(addr *Address, index int, count int) (string, error) {
	var b []*api.AccountBlock
	err := vc.c.Call(&b, "ledger_getBlocksByAccAddr", addr.address, index, count)
	if err != nil {
		return "", nil
	}
	jsonb, err := json.Marshal(b)
	if err != nil {
		return "", nil
	}
	return string(jsonb), nil
}

func (vc *Client) GetAccountByAccAddr(addr *Address) (string, error) {
	info := json.RawMessage{}
	err := vc.c.Call(&info, "ledger_getAccountByAccAddr", addr.address)
	if err != nil {
		return "", err
	}
	bytes, e := info.MarshalJSON()
	if e != nil {
		return "", e
	}
	return string(bytes), nil
}

func (vc *Client) GetOnroadAccountByAccAddr(addr *Address) (string, error) {
	info := json.RawMessage{}
	err := vc.c.Call(&info, "onroad_getAccountOnroadInfo", addr.address)
	if err != nil {
		return "", err
	}
	bytes, e := info.MarshalJSON()
	if e != nil {
		return "", e
	}
	return string(bytes), nil
}

func (vc *Client) GetSnapshotChainHeight() (string, error) {
	height := ""
	err := vc.c.Call(&height, "ledger_getSnapshotChainHeight")
	return height, err
}
