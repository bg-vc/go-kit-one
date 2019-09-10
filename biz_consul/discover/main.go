package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	consulHost = flag.String("consul.host", "localhost", "consul server ip address")
	consulPort = flag.String("consul.port", "8500", "consul server port")
)

func main() {
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var client consul.Client
	{
		consulCfg := api.DefaultConfig()
		consulCfg.Address = "http://" + *consulHost + ":" + *consulPort
		consulClient, err := api.NewClient(consulCfg)

		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}

		client = consul.NewClient(consulClient)
	}

	ctx := context.Background()

	discoverEndpoint := MakeDiscoverEndpoint(ctx, client, logger)

	r := MakeHttpHandler(discoverEndpoint, logger)

	errChan := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "http", "addr", "8002")
		handler := r
		errChan <- http.ListenAndServe(":8002", handler)
	}()

	logger.Log("exit", <-errChan)
}
