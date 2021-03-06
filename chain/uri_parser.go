package chain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iotexproject/assets/chain/contracts"
	"github.com/iotexproject/iotex-address/address"
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
	if info.TokenURI == "iotex_token_metadata" {
		resp, err := http.Get("https://raw.githubusercontent.com/iotexproject/iotex-token-metadata/master/token-metadata.json")
		if err != nil {
			return "", fmt.Errorf("fetch iotex token metadata error: %v", err)
		}
		defer resp.Body.Close()
		metadata, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read iotex token metadata body error: %v", err)
		}

		var details map[string]struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Logo        string   `json:"logo"`
			Type        string   `json:"type"`
			Symbol      string   `json:"symbol"`
			ImageUrls   []string `json:"image_urls"`
		}
		if err := json.Unmarshal(metadata, &details); err != nil {
			return "", fmt.Errorf("unmarshal iotex token metadata error: %v", err)
		}

		ethAddr := common.HexToAddress(info.Id)
		ioAddr, err := address.FromBytes(ethAddr.Bytes())
		if err != nil {
			return "", fmt.Errorf("convert io address error: %v", err)
		}
		if _, ok := details[ioAddr.String()]; ok {
			image = details[ioAddr.String()].ImageUrls[0]
		} else {
			return "", fmt.Errorf("can't found %s token metadata", ioAddr.String())
		}
	} else if info.TokenURI == "tokenURI" {
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
		tokenURL, err := contract.TokenURI(nil, tokenId)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		image = string(tokenURL)
	} else if strings.HasPrefix(info.TokenURI, "http_json_metadata") {
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
	} else if strings.HasPrefix(info.TokenURI, "ipfs_json_metadata") {
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
		metadataURL = strings.Replace(metadataURL, "ipfs://", "https://ipfs.io/ipfs/", 1)
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
		image = strings.Replace(image, "ipfs://", "https://ipfs.io/ipfs/", 1)
	}
	CACHE.Set("iotex:"+info.Id+":"+id, image, time.Minute*5)
	return image, nil
}
