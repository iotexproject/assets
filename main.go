package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/iotexproject/assets/chain"
	"github.com/iotexproject/assets/own"
)

func main() {
	chains := make(map[string]string)
	chains["1"] = "ethereum"
	chains["4689"] = "iotex"
	chains["4690"] = "iotex_test"

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("iotex assets")
	})

	// GET /tokenlist/:chain
	app.Get("/tokenlist/:chain", func(c *fiber.Ctx) error {
		chain := c.Params("chain")
		if strings.HasPrefix(chain, "/") || strings.HasPrefix(chain, "./") || strings.HasPrefix(chain, "..") {
			return c.Status(http.StatusBadRequest).SendString("forbid prefix")
		}
		if t, ok := chains[chain]; ok {
			chain = t
		}
		data, err := os.ReadFile("./blockchains/" + chain + "/tokenlist.json")
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("unsupported chain")
		}
		c.Context().Response.Header.SetContentType(fiber.MIMEApplicationJSON)

		return c.SendString(string(data))
	})

	// GET /token/:chain/:address
	app.Get("/token/:chain/:address", func(c *fiber.Ctx) error {
		chainName := c.Params("chain")
		if strings.HasPrefix(chainName, "/") || strings.HasPrefix(chainName, "./") || strings.HasPrefix(chainName, "..") {
			return c.Status(http.StatusBadRequest).SendString("forbid prefix")
		}
		if t, ok := chains[chainName]; ok {
			chainName = t
		}
		address := strings.ToLower(c.Params("address"))

		var data []byte

		if _, err := os.Stat("./blockchains/" + chainName + "/assets/" + address + "/info.json"); os.IsNotExist(err) {
			data, err = os.ReadFile("./blockchains/" + chainName + "/tokenlist.json")
			if err != nil {
				return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
			}
			var tokenList chain.TokenList
			if err = json.Unmarshal(data, &tokenList); err != nil {
				return c.Status(http.StatusInternalServerError).SendString("token info json error")
			}
			info, err := tokenList.ConvertDetialToInfo(address)
			if err != nil {
				return c.Status(http.StatusBadRequest).SendString("unsupported token")
			}
			data, _ = json.Marshal(info)
		} else {
			data, err = os.ReadFile("./blockchains/" + chainName + "/assets/" + address + "/info.json")
			if err != nil {
				return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
			}
		}

		c.Context().Response.Header.SetContentType(fiber.MIMEApplicationJSON)
		return c.SendString(string(data))
	})

	// GET /token/:chain/:address/image/:tokenId
	app.Get("token/:chain/:address/image/:tokenId", func(c *fiber.Ctx) error {
		chainName := c.Params("chain")
		if strings.HasPrefix(chainName, "/") || strings.HasPrefix(chainName, "./") || strings.HasPrefix(chainName, "..") {
			return c.Status(http.StatusBadRequest).SendString("forbid prefix")
		}
		if t, ok := chains[chainName]; ok {
			chainName = t
		}
		address := strings.ToLower(c.Params("address"))
		tokenId := c.Params("tokenId")

		if chainName == "iotex" {
			return c.SendString(fmt.Sprintf("https://image.mimo.exchange/%s/%s.png?width=300", address, tokenId))
		}

		var tokenInfo chain.TokenInfo
		var tokenList chain.TokenList

		data, err := os.ReadFile("./blockchains/" + chainName + "/tokenlist.json")
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
		}
		if err = json.Unmarshal(data, &tokenList); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("token info json error")
		}

		if _, err := os.Stat("./blockchains/" + chainName + "/assets/" + address + "/info.json"); os.IsNotExist(err) {
			info, err := tokenList.ConvertDetialToInfo(address)
			if err != nil {
				info = &chain.TokenInfo{
					Id:       address,
					TokenURI: "dynamic",
				}
				// process dynamic
				// return c.Status(http.StatusBadRequest).SendString("unsupported token")
			}
			tokenInfo = *info
		} else {
			data, err := os.ReadFile("./blockchains/" + chainName + "/assets/" + address + "/info.json")
			if err != nil {
				return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
			}
			if err := json.Unmarshal(data, &tokenInfo); err != nil {
				return c.Status(http.StatusInternalServerError).SendString("token info json error")
			}
		}

		rpc, err := tokenList.GetRPC()
		if err != nil {
			log.Printf("parse token image error: %v\n", err)
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}
		image, err := chain.ParseNFTImage(chainName, rpc, &tokenInfo, tokenId)
		if err != nil {
			log.Printf("parse token image error: %v\n", err)
			return c.Status(http.StatusInternalServerError).SendString("parse token image error")
		}
		return c.SendString(image)
	})

	app.Get("image/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		data, ok := chain.IMAGE_CACHE.Get(id)
		if !ok {
			log.Printf("fetch data from cache fail")
			return c.Status(http.StatusInternalServerError).SendString("fetch data from cache fail")
		}
		c.Response().Header.Add("Content-Type", "image/svg+xml")
		return c.Send(data.([]byte))
	})
	app.Get("image/static/:file", func(c *fiber.Ctx) error {
		file := c.Params("file")
		return c.SendFile("static/" + file)
	})

	app.Get("/account/:chain/own/:account", func(c *fiber.Ctx) error {
		chainName := c.Params("chain")
		if strings.HasPrefix(chainName, "/") || strings.HasPrefix(chainName, "./") || strings.HasPrefix(chainName, "..") {
			return c.Status(http.StatusBadRequest).SendString("forbid prefix")
		}
		if t, ok := chains[chainName]; ok {
			chainName = t
		}

		var tokenList chain.TokenList
		tokenData, err := os.ReadFile("./blockchains/" + chainName + "/tokenlist.json")
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
		}
		if err = json.Unmarshal(tokenData, &tokenList); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("token info json error")
		}

		account := c.Params("account")
		tokenType := c.Query("type", "721")
		skip, _ := strconv.Atoi(c.Query("skip", "0"))
		first, _ := strconv.Atoi(c.Query("first", "10"))
		var fetcher own.Fetcher
		rpc, err := tokenList.GetRPC()
		if err != nil {
			log.Printf("fetch rpc error: %v\n", err)
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}
		if chainName == "ethereum" {
			fetcher, err = own.NewEthereumFetcher(rpc)
			if err != nil {
				log.Printf("instance ethereum fetcher error: %v\n", err)
				return c.Status(http.StatusInternalServerError).SendString(err.Error())
			}
		} else if chainName == "iotex" {
			fetcher, err = own.NewIoTeXFetcher(rpc)
			if err != nil {
				log.Printf("instance iotex fetcher error: %v\n", err)
				return c.Status(http.StatusInternalServerError).SendString(err.Error())
			}
		} else if chainName == "iotex_test" {
			fetcher, err = own.NewIoTeXTestnetFetcher(rpc)
			if err != nil {
				log.Printf("instance iotex testnet fetcher error: %v\n", err)
				return c.Status(http.StatusInternalServerError).SendString(err.Error())
			}
		} else {
			return c.Status(http.StatusInternalServerError).SendString("chain does not supported")
		}

		data, err := fetcher.FetchOwnTokens(account, tokenType, skip, first)
		if err != nil {
			log.Printf("fetch own tokens error: %v\n", err)
			return c.Status(http.StatusInternalServerError).SendString("fetch own tokens error")
		}

		// process name and symbol with tokenlist
		for i, token := range data {
			for _, t := range tokenList.Tokens {
				if strings.EqualFold(token.Contract, t.Address) {
					data[i].Name = t.Name
					data[i].Symbol = t.Symbol
					break
				}
			}
		}

		return c.JSON(data)
	})

	log.Fatal(app.Listen(":3000"))
}
