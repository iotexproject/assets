package chain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iotexproject/assets/chain/contracts"
	"github.com/iotexproject/iotex-address/address"
	"github.com/vincent-petithory/dataurl"
)

func fetchTokenURI(endpoint string, info *TokenInfo, id string) (string, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return "", fmt.Errorf("connect rpc error: %v", err)
	}
	contractAddr := common.HexToAddress(info.Id)

	contract721, err := contracts.NewERC721(contractAddr, client)
	if err != nil {
		return "", fmt.Errorf("construct contract error: %v", err)
	}
	tokenId, _ := new(big.Int).SetString(id, 10)
	tokenURI, err := contract721.TokenURI(nil, tokenId)
	if err == nil {
		return tokenURI, nil
	}

	contract1155, err := contracts.NewERC1155(contractAddr, client)
	if err != nil {
		return "", fmt.Errorf("construct contract error: %v", err)
	}
	return contract1155.Uri(nil, tokenId)
}

func ParseNFTImage(network, endpoint string, info *TokenInfo, id string) (string, error) {
	if info.Type == "ERC20" {
		return "", nil
	}
	ci, found := CACHE.Get(network + ":" + info.Id + ":" + id)
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
		metadata, err := io.ReadAll(resp.Body)
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
		tokenURL, err := fetchTokenURI(endpoint, info, id)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		image = string(tokenURL)
	} else if strings.HasPrefix(info.TokenURI, "http_json_metadata") {
		metadataURL, err := fetchTokenURI(endpoint, info, id)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		segments := strings.Split(info.TokenURI, "_")
		image, err = parseHttpJsonMetadata(metadataURL, segments[3])
		if err != nil {
			return "", fmt.Errorf("parse image error: %v", err)
		}
	} else if strings.HasPrefix(info.TokenURI, "ipfs_json_metadata") {
		metadataURL, err := fetchTokenURI(endpoint, info, id)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		metadataURL = strings.Replace(metadataURL, "ipfs://", "https://ipfs.io/ipfs/", 1)
		resp, err := http.Get(metadataURL)
		if err != nil {
			return "", fmt.Errorf("fetch metadata error: %v", err)
		}
		defer resp.Body.Close()
		metadata, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read metadata body error: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(string(metadata)), &data); err != nil {
			return "", fmt.Errorf("unmarshal metadata error: %v", err)
		}
		segments := strings.Split(info.TokenURI, "_")
		image = data[segments[3]].(string)
		image = strings.Replace(image, "ipfs://", "https://ipfs.io/ipfs/", 1)
	} else if strings.HasPrefix(info.TokenURI, "data_json_metadata") {
		metadataURI, err := fetchTokenURI(endpoint, info, id)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		image, err = parseDataJsonMetedata(network, info.Id, metadataURI, info.TokenURI[19:], id)
		if err != nil {
			return "", fmt.Errorf("parse image error: %v", err)
		}
	} else if strings.HasPrefix(info.TokenURI, "static_replace") {
		idReg := regexp.MustCompile(`({(.*)})`)
		params := idReg.FindStringSubmatch(info.Template)
		if len(params) != 3 {
			return "", errors.New("error pattern")
		}
		image = strings.Replace(info.Template, params[0], fmt.Sprintf(params[2], id), 1)
	} else if strings.HasPrefix(info.TokenURI, "static") {
		image = os.Getenv("SITE_URL") + "/image/static/" + info.TokenURI[7:]
	} else if strings.HasPrefix(info.TokenURI, "ar_json_metadata") {
		metadataURL, err := fetchTokenURI(endpoint, info, id)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		segments := strings.Split(info.TokenURI, "_")
		image, err = parseArJsonMetadata(metadataURL, segments[3])
		if err != nil {
			return "", fmt.Errorf("parse image error: %v", err)
		}
	} else {
		tokenURL, err := fetchTokenURI(endpoint, info, id)
		if err != nil {
			return "", fmt.Errorf("read tokenURI error: %v", err)
		}
		if strings.HasPrefix(tokenURL, "data:application/json;base64,") {
			image, err = parseDataJsonMetedata(network, info.Id, tokenURL, "image", id)
			if err != nil {
				fmt.Println(err)
				return "", fmt.Errorf("parse image error")
			}
		} else if strings.HasPrefix(tokenURL, "http://") || strings.HasPrefix(tokenURL, "https://") {
			image, err = parseHttpJsonMetadata(tokenURL, "image")
			if err != nil {
				return "", fmt.Errorf("parse image error")
			}
		} else if strings.HasPrefix(tokenURL, "ipfs://") {
			tokenURL = strings.Replace(tokenURL, "ipfs://", "https://ipfs.io/ipfs/", 1)
			image, err = parseHttpJsonMetadata(tokenURL, "image")
			if err != nil {
				return "", fmt.Errorf("parse image error")
			}
		} else if strings.HasPrefix(tokenURL, "ipfs://") {
			image, err = parseArJsonMetadata(tokenURL, "image")
			if err != nil {
				return "", fmt.Errorf("parse image error")
			}
		} else {
			return "", fmt.Errorf("unsupported token")
		}
	}
	CACHE.Set(network+":"+info.Id+":"+id, image, time.Minute*5)
	return image, nil
}

func parseDataJsonMetedata(network, address, metadataURI, imageFieldName, id string) (string, error) {
	metadata, err := dataurl.DecodeString(metadataURI)
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(metadata.Data, &data); err != nil {
		return "", fmt.Errorf("unmarshal metadata error: %v", err)
	}
	imageField := data[imageFieldName].(string)
	if strings.HasPrefix(imageField, "http") {
		return imageField, nil
	}
	imageData, err := dataurl.DecodeString(imageField)
	if err != nil {
		return "", err
	}
	imageKey := network + "_" + address + "_" + id
	image := os.Getenv("SITE_URL") + "/image/" + imageKey
	IMAGE_CACHE.Set(imageKey, imageData.Data, time.Minute*10)

	return image, nil
}

func parseHttpJsonMetadata(tokenURI, imageFieldName string) (string, error) {
	resp, err := http.Get(tokenURI)
	if err != nil {
		return "", fmt.Errorf("fetch metadata error: %v", err)
	}
	defer resp.Body.Close()
	metadata, err := io.ReadAll(resp.Body)
	metadata = bytes.TrimPrefix(metadata, []byte("\xef\xbb\xbf"))
	if err != nil {
		return "", fmt.Errorf("read metadata body error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(string(metadata)), &data); err != nil {
		return "", fmt.Errorf("unmarshal metadata error: %v", err)
	}
	image := data[imageFieldName].(string)
	if strings.HasPrefix(image, "ipfs://") {
		image = strings.Replace(image, "ipfs://", "https://ipfs.io/ipfs/", 1)
	}
	if strings.HasPrefix(image, "ar://") {
		image = strings.Replace(image, "ar://", "https://arweave.net/", 1)
	}
	return image, nil
}

func parseArJsonMetadata(tokenURI, imageFieldName string) (string, error) {
	tokenURI = strings.Replace(tokenURI, "ar://", "https://arweave.net/", 1)
	resp, err := http.Get(tokenURI)
	if err != nil {
		return "", fmt.Errorf("fetch metadata error: %v", err)
	}
	defer resp.Body.Close()
	metadata, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read metadata body error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(metadata, &data); err != nil {
		return "", fmt.Errorf("unmarshal metadata error: %v", err)
	}
	image := data[imageFieldName].(string)
	if strings.HasPrefix(image, "ipfs://") {
		image = strings.Replace(image, "ipfs://", "https://ipfs.io/ipfs/", 1)
	}
	if strings.HasPrefix(image, "ar://") {
		image = strings.Replace(image, "ar://", "https://arweave.net/", 1)
	}
	return image, nil
}
