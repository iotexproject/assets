package chain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iotexproject/assets/chain/contracts"
	"github.com/patrickmn/go-cache"
)

func ParseNFTImage(info *TokenInfo, id string) (string, error) {
	if info.Type == "ERC20" {
		return "", nil
	}
	ci, found := CACHE.Get("iotex:" + info.Id + ":" + id)
	if found {
		return ci.(string), nil
	}
	var image string
	if strings.HasPrefix(info.TokenURI, "http_json_metadata") {
		client, err := ethclient.Dial("https://babel-api.mainnet.iotex.io/")
		if err != nil {
			return "", fmt.Errorf("connect rpc error: %v", err)
		}
		contractAddr := common.HexToAddress(info.Id)

		contract, err := contracts.NewERC721(contractAddr, client)
		if err != nil {
			return "", fmt.Errorf("construct contract error: %v", err)
		}
		tokenId, _ := new(big.Int).SetString(id, 10)
		metadataURL, err := contract.TokenURI(nil, tokenId)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		resp, err := http.Get(metadataURL)
		if err != nil {
			return "", fmt.Errorf("fetch metadata error: %v", err)
		}
		defer resp.Body.Close()
		metadata, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read metadata body error: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(metadata, &data); err != nil {
			return "", fmt.Errorf("unmarshal metadata error: %v", err)
		}
		segments := strings.Split(info.TokenURI, "_")
		image = data[segments[3]].(string)
	}
	CACHE.Set("iotex:"+info.Id+":"+id, image, cache.DefaultExpiration)
	return image, nil
}
