package own

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hasura/go-graphql-client"
	"github.com/iotexproject/assets/chain/contracts"
)

type OwnToken struct {
	Contract string `json:"contract"`
	Type     string `json:"type"`
	TokenId  string `json:"tokenId"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Amount   string `json:"amount"`
	SBT      bool   `json:"sbt"`
}

type Fetcher interface {
	FetchOwnTokens(account string, tokenType string, skip int, first int) ([]OwnToken, error)
}

type EthereumFetcher struct {
	client *graphql.Client
	rpc    *ethclient.Client
}

func TryCheckSBT(rpc *ethclient.Client, contract string) (bool, error) {
	contractAddr := common.HexToAddress(contract)

	contract721, err := contracts.NewERC165(contractAddr, rpc)
	if err != nil {
		return false, fmt.Errorf("construct contract error: %v", err)
	}
	interfaceId := [4]byte{180, 90, 60, 14}
	sbt, err := contract721.SupportsInterface(nil, interfaceId)
	if err == nil {
		return sbt, nil
	}
	return false, nil
}

func NewEthereumFetcher(endpoint string) (*EthereumFetcher, error) {
	client := graphql.NewClient("https://api.thegraph.com/subgraphs/name/wighawag/eip721-subgraph", nil)
	rpc, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("connect rpc error: %v", err)
	}
	return &EthereumFetcher{client: client, rpc: rpc}, nil
}

func (f *EthereumFetcher) FetchOwnTokens(account string, tokenType string, skip int, first int) ([]OwnToken, error) {
	var q struct {
		Tokens []struct {
			Id       string
			TokenID  string `graphql:"tokenID"`
			Contract struct {
				Name   string
				Symbol string
			}
		} `graphql:"tokens(skip: $skip first: $first where: {owner: $owner})"`
	}

	variables := map[string]interface{}{
		"owner": graphql.ID(account),
		"skip":  skip,
		"first": first,
	}

	err := f.client.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, err
	}

	result := make([]OwnToken, len(q.Tokens))
	for i, token := range q.Tokens {
		sbt, err := TryCheckSBT(f.rpc, token.Id[:42])
		if err != nil {
			return nil, err
		}
		result[i] = OwnToken{
			Contract: token.Id[:42],
			Type:     tokenType,
			TokenId:  token.TokenID,
			Name:     token.Contract.Name,
			Symbol:   token.Contract.Symbol,
			SBT:      sbt,
		}
	}

	return result, nil
}
