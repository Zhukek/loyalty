package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Zhukek/loyalty/internal/client"
	"github.com/Zhukek/loyalty/internal/handler"
	"github.com/Zhukek/loyalty/internal/logger"
	"github.com/Zhukek/loyalty/internal/logger/slogger"
	"github.com/Zhukek/loyalty/internal/repository/postgresql"
)

func main() {

	logger, err := slogger.NewSlogger()

	if err != nil {
		fmt.Printf("logger start err: %s", err)
		return
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
		DBURI          = config.DBURI
		accrual        = config.AccrualAddress
	)

	if DBURI == "" {
		return fmt.Errorf("no DB URI")
	}

	rep, err := postgresql.NewPGRepository(DBURI)

	if err != nil {
		return err
	}

	defer rep.Close()

	client := client.NewtClient(accrual, rep, logger)
	defer client.Close()

	router := handler.NewRouter(logger, rep)

	fmt.Printf("Running on %s\n", address)
	return http.ListenAndServe(address, router)
}
