package main

import (
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	zipkinHttpsvr "github.com/openzipkin/zipkin-go/middleware/http"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

type HystrixRouter struct {
	svcMap       *sync.Map
	logger       log.Logger
	fallbackMsg  string
	consulClient *api.Client
	tracer       *zipkin.Tracer
}

func NewRoutes(client *api.Client, tracer *zipkin.Tracer, fbMsg string, logger log.Logger) http.Handler {
	return &HystrixRouter{
		svcMap:       &sync.Map{},
		logger:       logger,
		fallbackMsg:  fbMsg,
		consulClient: client,
		tracer:       tracer,
	}
}

func (router *HystrixRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /biz/add/1/2
	reqPath := r.URL.Path
	if reqPath == "" {
		return
	}

	pathArray := strings.Split(reqPath, "/")
	serviceName := pathArray[1]

	if _, ok := router.svcMap.Load(serviceName); !ok {
		hystrix.ConfigureCommand(serviceName, hystrix.CommandConfig{Timeout: 1000})
		router.svcMap.Store(serviceName, serviceName)
	}

	err := hystrix.Do(serviceName, func() (err error) {
		result, _, err := router.consulClient.Catalog().Service(serviceName, "", nil)
		if err != nil {
			router.logger.Log("reverseProxy failed", "query service instance error", err.Error())
			return
		}

		if len(result) == 0 {
			router.logger.Log("reverseProxy failed", "no such service instance", serviceName)
			return errors.New("no such service instance")
		}

		director := func(req *http.Request) {
			destPath := strings.Join(pathArray[2:], "/")
			tgt := result[rand.Int()%len(result)]
			router.logger.Log("service id", tgt.ServiceID)

			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf("%s:%d", tgt.ServiceAddress, tgt.ServicePort)
			req.URL.Path = "/" + destPath
		}

		var proxyError error = nil
		roundTrip, _ := zipkinHttpsvr.NewTransport(router.tracer, zipkinHttpsvr.TransportTrace(true))

		errorHandler := func(ew http.ResponseWriter, er *http.Request, err error) {
			proxyError = err
		}
		proxy := &httputil.ReverseProxy{
			Director:     director,
			Transport:    roundTrip,
			ErrorHandler: errorHandler,
		}

		proxy.ServeHTTP(w, r)
		return proxyError
	}, func(err error) error {
		router.logger.Log("fallback error desc", err.Error())
		return errors.New(router.fallbackMsg)
	})

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}
