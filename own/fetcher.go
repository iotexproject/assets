package own

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type OwnToken struct {
	Contract string `json:"contract"`
	TokenId  string `json:"tokenId"`
}

type SubgraphFetcher struct {
	client *graphql.Client
}

func NewEthereumFetcher() *SubgraphFetcher {
	client := graphql.NewClient("https://api.thegraph.com/subgraphs/name/wighawag/eip721-subgraph", nil)
	return &SubgraphFetcher{client: client}
}

func NewIoTeXFetcher() *SubgraphFetcher {
	client := graphql.NewClient("https://graph.mainnet.iotex.io/subgraphs/name/ququzone/eip721", nil)
	return &SubgraphFetcher{client: client}
}

func (f *SubgraphFetcher) FetchOwnTokens(account string, skip int, first int) ([]OwnToken, error) {
	var q struct {
		Tokens []struct {
			Id      string
			TokenID string `graphql:"tokenID"`
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
			TokenId:  token.TokenID,
		}
	}

	return result, nil
}
