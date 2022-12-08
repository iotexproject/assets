package own

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type IoTeXFetcher struct {
	client *graphql.Client
}

func NewIoTeXFetcher() *IoTeXFetcher {
	client := graphql.NewClient("https://graph.mainnet.iotex.io/subgraphs/name/ququzone/eip721", nil)
	return &IoTeXFetcher{client: client}
}

func (f *IoTeXFetcher) FetchOwnTokens(account string, skip int, first int) ([]OwnToken, error) {
	var q struct {
		Tokens []struct {
			Id         string
			TokenID    string `graphql:"tokenID"`
			Collection struct {
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
			TokenId:  token.TokenID,
			Name:     token.Collection.Name,
			Symbol:   token.Collection.Symbol,
		}
	}

	return result, nil
}
