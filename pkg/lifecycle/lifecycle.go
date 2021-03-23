package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/log"
)

// TerminateHandler terminates the application.
// This function waits on a SIGINT or SIGTERM signal and shuts down the HTTP servers gracefully.
func TerminateHandler(logger log.Logger, server1 core.ShutDowner, server2 core.ShutDowner) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down application ...")

	// We will wait 5 seconds for the server to shutdown gracefully
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	err := server1.ShutDown(ctx1)
	if err != nil {
		logger.Error(fmt.Sprintf("api server1 failed to shutdown gracefully: %s", err.Error()))
	}

	// We will wait 5 seconds for the server to shutdown gracefully
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	err = server2.ShutDown(ctx2)
	if err != nil {
		logger.Error(fmt.Sprintf("api server2 failed to shutdown gracefully: %s", err.Error()))
	}
}
