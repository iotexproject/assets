package own

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type IoTeXFetcher struct {
	client721  *graphql.Client
	client1155 *graphql.Client
}

func NewIoTeXFetcher() *IoTeXFetcher {
	client721 := graphql.NewClient("https://graph.mainnet.iotex.io/subgraphs/name/ququzone/eip721", nil)
	client1155 := graphql.NewClient("https://graph.mainnet.iotex.io/subgraphs/name/ququzone/eip1155", nil)
	return &IoTeXFetcher{client721: client721, client1155: client1155}
}

func NewIoTeXTestnetFetcher() *IoTeXFetcher {
	client721 := graphql.NewClient("https://graph.mainnet.iotex.io/subgraphs/name/looksrare/eip721", nil)
	client1155 := graphql.NewClient("https://graph.mainnet.iotex.io/subgraphs/name/testnet/eip1155", nil)
	return &IoTeXFetcher{client721: client721, client1155: client1155}
}

func (f *IoTeXFetcher) fetch721(account string, skip int, first int) ([]OwnToken, error) {
	var q struct {
		Tokens []struct {
			Id         string
			TokenID    string `graphql:"tokenID"`
			Collection struct {
				Name   string
				Symbol string
			}
		} `graphql:"tokens(skip: $skip first: $first where: {collection_not: \"0xec0cd5c1d61943a195bca7b381dc60f9f545a540\" owner: $owner})"`
	}

	variables := map[string]interface{}{
		"owner": graphql.ID(account),
		"skip":  skip,
		"first": first,
	}

	err := f.client721.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, err
	}

	result := make([]OwnToken, len(q.Tokens))
	for i, token := range q.Tokens {
		result[i] = OwnToken{
			Contract: token.Id[:42],
			Type:     "721",
			TokenId:  token.TokenID,
			Name:     token.Collection.Name,
			Symbol:   token.Collection.Symbol,
			Amount:   "1",
		}
	}

	return result, nil
}

func tryFetchName(tokenURI string) (string, error) {
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
	name, ok := data["name"]
	if !ok {
		return "", errors.New("can't find name field")
	}
	return name.(string), nil
}

func (f *IoTeXFetcher) fetch1155(account string, skip int, first int) ([]OwnToken, error) {
	var q struct {
		TokenOwners []struct {
			Id     string
			Amount string
			Token  struct {
				TokenID    string `graphql:"tokenID"`
				TokenURI   string `graphql:"tokenURI"`
				Collection struct {
					Name   string
					Symbol string
				}
			}
		} `graphql:"tokenOwners(skip: $skip first: $first where: {owner: $owner amount_gt: 0})"`
	}

	variables := map[string]interface{}{
		"owner": graphql.ID(account),
		"skip":  skip,
		"first": first,
	}

	err := f.client1155.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, err
	}

	result := make([]OwnToken, len(q.TokenOwners))
	for i, token := range q.TokenOwners {
		result[i] = OwnToken{
			Contract: token.Id[:42],
			Type:     "1155",
			TokenId:  token.Token.TokenID,
			Name:     token.Token.Collection.Name,
			Symbol:   token.Token.Collection.Symbol,
			Amount:   token.Amount,
		}
		if result[i].Name == "unknown" {
			tokenURI := token.Token.TokenURI
			if tokenURI != "" {
				name, err := tryFetchName(tokenURI)
				if err == nil {
					result[i].Name = name
				}
			}
		}
	}

	return result, nil
}

func (f *IoTeXFetcher) FetchOwnTokens(account string, tokenType string, skip int, first int) ([]OwnToken, error) {
	switch tokenType {
	case "721":
		return f.fetch721(account, skip, first)
	case "1155":
		return f.fetch1155(account, skip, first)
	default:
		return nil, errors.New("unsupport type")
	}
}
