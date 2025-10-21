package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Zhukek/loyalty/internal/logger"
	"github.com/Zhukek/loyalty/internal/models"
	"github.com/Zhukek/loyalty/internal/repository"
	"resty.dev/v3"
)

type Client struct {
	c       *resty.Client
	rep     repository.Repository
	accrual string
	logger  logger.Logger
	jobs    chan models.Order
}

func (c *Client) addJobs() {
	orders, err := c.rep.GetProcessingOrders(context.Background())
	if err != nil {
		c.logger.LogErr("client get orders", err)
		return
	}

	for _, order := range orders {
		c.jobs <- order
	}
}

func (c *Client) worker() {
	for order := range c.jobs {
		var resOrder models.AccrualOrder

		resp, err := c.c.R().
			SetPathParam("orderID", order.Number).
			SetHeader("Accept", "application/json").
			SetResult(&resOrder).
			Get(c.accrual + "/api/orders/{orderID}")

		if err != nil {
			c.logger.LogErr("get accrual", err)
			continue
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			fmt.Println(resOrder.Order)
			/// update
		case http.StatusNoContent:
			/// invalid
			continue
		case http.StatusTooManyRequests:
			time.Sleep(1 * time.Minute)

		}
	}
}

func (c *Client) start() {
	numWorkers := 3
	ticker := time.NewTicker(5 * time.Second)

	for i := 1; i <= numWorkers; i++ {
		go c.worker()
	}

	for {
		<-ticker.C
		if len(c.jobs) == 0 {
			c.addJobs()
		}
	}
}

func (c *Client) Close() {
	c.c.Close()
	close(c.jobs)
}

func NewtClient(accrualAddress string, rep repository.Repository, logger logger.Logger) *Client {
	restyClient := resty.New()
	numJobs := 6
	jobsChan := make(chan models.Order, numJobs)

	client := Client{
		c:       restyClient,
		rep:     rep,
		accrual: accrualAddress,
		logger:  logger,
		jobs:    jobsChan,
	}

	go client.start()

	return &client
}
