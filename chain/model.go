package chain

import (
	"errors"
	"os"
	"strings"
)

type TokenList struct {
	Name      string `json:"name"`
	LogoURI   string `json:"logoURI"`
	Timestamp string `json:"timestamp"`
	ChainId   int    `json:"chainId"`
	RPC       string `json:"rpc"`
	Symbol    string `json:"symbol"`
	Tokens    []struct {
		Type     string `json:"type"`
		Address  string `json:"address"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		TokenURI string `json:"tokenURI"`
		Template string `json:"template"`
		LogoURI  string `json:"logoURI"`
	} `json:"tokens"`
}

type TokenInfo struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Website     string `json:"website"`
	Description string `json:"description"`
	Explorer    string `json:"explorer"`
	Type        string `json:"type"`
	TokenURI    string `json:"tokenURI"`
	Symbol      string `json:"symbol"`
	Decimals    uint   `json:"decimals"`
	Status      string `json:"status"`
	Template    string `json:"template"`
}

func (tl *TokenList) ConvertDetialToInfo(id string) (*TokenInfo, error) {
	for _, v := range tl.Tokens {
		if strings.EqualFold(id, v.Address) {
			return &TokenInfo{
				Id:       strings.ToLower(id),
				Name:     v.Name,
				Type:     v.Type,
				Status:   "active",
				TokenURI: v.TokenURI,
				Template: v.Template,
			}, nil
		}
	}
	return nil, errors.New("unsupported token")
}

func (tl *TokenList) GetRPC() (string, error) {
	if strings.Contains(tl.RPC, "${KEY}") {
		key := os.Getenv("KEY")
		if key == "" {
			return "", errors.New("get rpc key error")
		}
		return strings.ReplaceAll(tl.RPC, "${KEY}", key), nil
	}
	return tl.RPC, nil
}
