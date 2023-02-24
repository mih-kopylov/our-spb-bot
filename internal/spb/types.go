package spb

type Message struct {
	categoryId  string
	fileUrls    []string
	description string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}
