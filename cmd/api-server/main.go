package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api/apimerchant"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api/apimgmt"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/log"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/lifecycle"
)

func main() {
	retCode := mainLogic()
	os.Exit(retCode)
}

func mainLogic() int {
	// Setup logger
	logger := core.NewAppLogger(os.Stdout, log.INFO)
	defer logger.Sync()

	logger.Info("APP starting")

	// Read config
	logger.Info("reading configuration", log.Field("type", "setup"))
	config := core.NewConfig()
	if err := config.LoadConfig(); err != nil {
		logger.Error(err.Error(), log.Field("type", "setup"))
		return 1
	}

	// TODO: Set log level after reading config
	// something like this:
	// logger.SetLevel(config.Options.LogLevel)

	// Setup Database

	serverMerchant := apimerchant.NewServer(config.WebserverMerchant.Host, config.WebserverMerchant.Port, config.Options.DevMode, logger, nil)
	serverMgmt := apimgmt.NewServer(config.WebserverMgmt.Host, config.WebserverMgmt.Port, config.Options.DevMode, logger, nil)

	// Spawn SIGINT listener
	go lifecycle.TerminateHandler(logger, serverMerchant, serverMgmt)

	var wg sync.WaitGroup
	wg.Add(2)

	go RunMerchantWebserver(logger, &wg, serverMerchant)
	go RunMgmtWebserver(logger, &wg, serverMgmt)

	// Wait here for both web servers to return
	wg.Wait()

	// Handle case when it crashes. Add a chan of ints or errors to return 1

	logger.Info("APP gracefully terminated")
	return 0
}

func RunMerchantWebserver(logger log.Logger, wg *sync.WaitGroup, serverMerchant *apimerchant.Server) {
	defer wg.Done()

	logger.Info("listenning for incoming requests", log.Field("type", "setup"))
	err := serverMerchant.ListenAndServe()
	if err != nil {
		logger.Error(fmt.Sprintf("unexpected error while serving HTTP: %s", err))
	}
}

func RunMgmtWebserver(logger log.Logger, wg *sync.WaitGroup, serverMgmt *apimgmt.Server) {
	defer wg.Done()

	logger.Info("listenning for incoming requests", log.Field("type", "setup"))
	err := serverMgmt.ListenAndServe()
	if err != nil {
		logger.Error(fmt.Sprintf("unexpected error while serving HTTP: %s", err))
	}
}
