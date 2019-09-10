package main

import (
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
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

	consulCfg := api.DefaultConfig()
	consulCfg.Address = "http://" + *consulHost + ":" + *consulPort
	consulClient, err := api.NewClient(consulCfg)

	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	proxy := NewReverseProxy(consulClient, logger)

	errChan := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "http", "addr", "8003")
		handler := proxy
		errChan <- http.ListenAndServe(":8003", handler)
	}()

	logger.Log("exit", <-errChan)
}

func NewReverseProxy(client *api.Client, logger log.Logger) *httputil.ReverseProxy{
	director := func(req *http.Request) {
		reqPath := req.URL.Path
		if reqPath == "" {
			return
		}
		// /biz/add/1/2
		pathArray := strings.Split(reqPath, "/")
		serviceName := pathArray[1]
		logger.Log("serviceName:", serviceName)

		result, _, err := client.Catalog().Service(serviceName, "", nil)
		if err != nil {
			logger.Log("reverseProxy failed", "query service instance error", err.Error())
			return
		}

		if len(result) == 0 {
			logger.Log("reverseProxy failed", "no such service instance", serviceName)
			return
		}

		destPath := strings.Join(pathArray[2:], "/")
		tgt := result[rand.Int()%len(result)]
		logger.Log("service id", tgt.ServiceID)

		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("%s:%d", tgt.ServiceAddress, tgt.ServicePort)
		req.URL.Path = "/" + destPath
	}

	return &httputil.ReverseProxy{Director:director}
}
