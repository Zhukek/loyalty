package client

import (
	"github.com/Zhukek/loyalty/internal/repository"
	"resty.dev/v3"
)

type Client struct {
	C       *resty.Client
	Rep     repository.Repository
	Accrual string
}

func (c *Client) Close() {
	c.C.Close()
}

func NewtClient(accrualAddress string, rep repository.Repository) *Client {
	client := resty.New()

	return &Client{
		C:       client,
		Rep:     rep,
		Accrual: accrualAddress,
	}
}
