package downloader

import (
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
)

type ProxyFetcher struct {
	d *Downloader
	urls []string
	client *http.Client
}

type getBlockByHashResponse struct {
	Jsonrpc    string           `json:"jsonrpc"`
	Result     *types.Header    `json:"result"`
	Error      interface{}      `json:"error"`
	Id         uint64           `json:"id"`
}

func NewProxyFetcher(d *Downloader, isMainnet bool) ProxyFetcher {
	pf := ProxyFetcher{d, nil, new(http.Client)}

	if isMainnet {
		pf.urls = []string{
			"https://bsc-dataseed.binance.org",
			"https://bsc-dataseed1.defibit.io",
			"https://bsc-dataseed1.ninicoin.io",
			"https://bsc-dataseed2.defibit.io",
			"https://bsc-dataseed3.defibit.io",
			"https://bsc-dataseed4.defibit.io",
			"https://bsc-dataseed2.ninicoin.io",
			"https://bsc-dataseed3.ninicoin.io",
			"https://bsc-dataseed4.ninicoin.io",
			"https://bsc-dataseed1.binance.org",
			"https://bsc-dataseed2.binance.org",
			"https://bsc-dataseed3.binance.org",
			"https://bsc-dataseed4.binance.org",
		}
	} else {
		pf.urls = []string{
			"https://data-seed-prebsc-1-s1.binance.org:8545",
			"https://data-seed-prebsc-2-s1.binance.org:8545",
			"https://data-seed-prebsc-1-s2.binance.org:8545",
			"https://data-seed-prebsc-2-s2.binance.org:8545",
			"https://data-seed-prebsc-1-s3.binance.org:8545",
			"https://data-seed-prebsc-2-s3.binance.org:8545",
		}
	}

	return pf
}

func (pf ProxyFetcher) GetHeaderByHash(hash common.Hash) *types.Header {
	var (
		id uint64
		header *types.Header
	)

	for {
		for _, url := range pf.urls {
			id++
			msg := fmt.Sprintf(`{"jsonrpc":"2.0", "method":"eth_getBlockByHash", "params":[%q,true], "id":%d}`, hash, id)
			req, err := http.NewRequest("POST", url, strings.NewReader(msg))
			if err != nil {
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := pf.client.Do(req)
			if err != nil {
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			res := new(getBlockByHashResponse)
			err = json.Unmarshal(body, res)
			if err == nil && res.Error == nil && res.Result != nil {
				header = res.Result
				break
			}
		}

		if header != nil {
			break
		}
	}

	return header
}

func (pf ProxyFetcher) GetHeader(db ethdb.Database, number uint64) *types.Header {
	log.Info("ProxyFetcher GetHeader", "number", number)

	hash := rawdb.ReadCanonicalHash(db, number)
	header := rawdb.ReadHeader(db, hash, number)

	if hash == (common.Hash{}) || header != nil {
		return nil
	}

	header = pf.GetHeaderByHash(hash)
    rawdb.WriteHeader(db, header)

	return header
}
