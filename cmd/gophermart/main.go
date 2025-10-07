package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Zhukek/loyalty/internal/handler"
	"github.com/Zhukek/loyalty/internal/logger"
	"github.com/Zhukek/loyalty/internal/logger/slogger"
)

func main() {

	logger, err := slogger.NewSlogger()

	if err != nil {
		fmt.Printf("logger start err: %s", err)
	}
	defer logger.Sync()

	if err := run(logger); err != nil {
		log.Fatal("run err:", err)
	}
}

func run(logger logger.Logger) error {
	config := getParams()

	var (
		address string = config.Address
		// DBURI          = config.DBURI
		// accrual        = config.AccrualAddress
	)

	router := handler.NewRouter(logger)

	fmt.Printf("Running on %s\n", address)
	return http.ListenAndServe(address, router)
}
