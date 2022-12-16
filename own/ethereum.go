package own

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type OwnToken struct {
	Contract string `json:"contract"`
	Type     string `json:"type"`
	TokenId  string `json:"tokenId"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Amount   string `json:"amount"`
}

type Fetcher interface {
	FetchOwnTokens(account string, tokenType string, skip int, first int) ([]OwnToken, error)
}

type EthereumFetcher struct {
	client *graphql.Client
}

func NewEthereumFetcher() *EthereumFetcher {
	client := graphql.NewClient("https://api.thegraph.com/subgraphs/name/wighawag/eip721-subgraph", nil)
	return &EthereumFetcher{client: client}
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
		result[i] = OwnToken{
			Contract: token.Id[:42],
			Type:     tokenType,
			TokenId:  token.TokenID,
			Name:     token.Contract.Name,
			Symbol:   token.Contract.Symbol,
		}
	}

	return result, nil
}
