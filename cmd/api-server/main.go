package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api/apimerchant"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api/apimgmt"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/log"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/pprocessor"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/repository"
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
	db, err := repository.NewDatabaseService(config.Database.Host, config.Database.Port,
		config.Database.Username, config.Database.Password, config.Database.DBName)
	if err != nil {
		logger.Error(fmt.Sprintf("database error: %s", err.Error()), log.Field("type", "setup"))
		return 1
	}
	defer db.Close()

	httpClient := &http.Client{
		Timeout: time.Second * time.Duration(config.Options.HTTPClientTimeout),
	}

	// Setup Payment processor service
	pprocservice := pprocessor.NewClient(config.PProcessorService.Host, config.PProcessorService.Port, httpClient)

	serverMerchant := apimerchant.NewServer(config.WebserverMerchant.Host, config.WebserverMerchant.Port, config.Options.DevMode,
		config.AuthService.Host, config.AuthService.Port,
		logger, httpClient, db, pprocservice)
	serverMgmt := apimgmt.NewServer(config.WebserverMgmt.Host, config.WebserverMgmt.Port, config.Options.DevMode, logger, db)

	// Spawn SIGINT listener
	go lifecycle.TerminateHandler(logger, serverMerchant, serverMgmt)

	errSignal := make(chan struct{}, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go RunMerchantWebserver(logger, serverMerchant, &wg, errSignal)
	go RunMgmtWebserver(logger, serverMgmt, &wg, errSignal)

	// Wait here for both web servers to return
	wg.Wait()

	select {
	case <-errSignal:
		return 1
	default:
		logger.Info("APP gracefully terminated")
		return 0
	}
}

func RunMerchantWebserver(logger log.Logger, serverMerchant *apimerchant.Server, wg *sync.WaitGroup, errSignal chan struct{}) {
	defer wg.Done()

	logger.Info("listenning for incoming requests on apimerchant", log.Field("type", "setup"))
	err := serverMerchant.ListenAndServe()
	if err != nil {
		logger.Error(fmt.Sprintf("unexpected error while serving HTTP on apimerchant: %s", err))
		errSignal <- struct{}{}
	}
}

func RunMgmtWebserver(logger log.Logger, serverMgmt *apimgmt.Server, wg *sync.WaitGroup, errSignal chan struct{}) {
	defer wg.Done()

	logger.Info("listenning for incoming requests on apimgmt", log.Field("type", "setup"))
	err := serverMgmt.ListenAndServe()
	if err != nil {
		logger.Error(fmt.Sprintf("unexpected error while serving HTTP on apimgmt: %s", err))
		errSignal <- struct{}{}
	}
}
