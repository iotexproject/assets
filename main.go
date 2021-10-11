package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/iotexproject/assets/chain"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("iotex assets")
	})

	// GET /tokenlist/:chain
	app.Get("/tokenlist/:chain", func(c *fiber.Ctx) error {
		chain := c.Params("chain")
		if strings.HasPrefix(chain, "/") || strings.HasPrefix(chain, "./") || strings.HasPrefix(chain, "..") {
			return c.Status(http.StatusBadRequest).SendString("forbid prefix")
		}
		data, err := ioutil.ReadFile("./blockchains/" + chain + "/tokenlist.json")
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("unsupported chain")
		}
		c.Context().Response.Header.SetContentType(fiber.MIMEApplicationJSON)

		return c.SendString(string(data))
	})

	// GET /token/:chain/:address
	app.Get("/token/:chain/:address", func(c *fiber.Ctx) error {
		chain := c.Params("chain")
		if strings.HasPrefix(chain, "/") || strings.HasPrefix(chain, "./") || strings.HasPrefix(chain, "..") {
			return c.Status(http.StatusBadRequest).SendString("forbid prefix")
		}
		address := strings.ToLower(c.Params("address"))
		data, err := ioutil.ReadFile("./blockchains/" + chain + "/assets/" + address + "/info.json")
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
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
		address := strings.ToLower(c.Params("address"))
		data, err := ioutil.ReadFile("./blockchains/" + chainName + "/assets/" + address + "/info.json")
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("unsupported chain or token")
		}

		var tokenInfo chain.TokenInfo
		if err := json.Unmarshal(data, &tokenInfo); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("token info json error")
		}
		image, err := chain.ParseNFTImage(&tokenInfo, c.Params("tokenId"))
		if err != nil {
			log.Printf("parse token image error: %v", err)
			return c.Status(http.StatusInternalServerError).SendString("parse token image error")
		}
		return c.SendString(image)
	})

	log.Fatal(app.Listen(":3000"))
}
