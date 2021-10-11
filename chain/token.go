package chain

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
}
