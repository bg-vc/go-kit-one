package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	kitPrometheus "github.com/go-kit/kit/metrics/prometheus"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	consulHost  = flag.String("consul.host", "localhost", "consul ip address")
	consulPort  = flag.String("consul.port", "8500", "consul port")
	serviceHost = flag.String("service.host", "192.168.0.103", "service ip address")
	servicePort = flag.String("service.port", "8000", "service port")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	fieldKeys := []string{"method"}
	requestCount := kitPrometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: "vince_cfl",
		Subsystem: "biz_service",
		Name:      "request_count",
		Help:      "numbers of request received",
	}, fieldKeys)

	requestLatency := kitPrometheus.NewSummaryFrom(stdPrometheus.SummaryOpts{
		Namespace: "vince_cfl",
		Subsystem: "biz_service",
		Name:      "request_latency",
		Help:      "total duration of request in microseconds",
	}, fieldKeys)

	svc := NewBizService()

	svc = NewLoggingMiddleware(logger)(svc)

	svc = NewMetrics(requestCount, requestLatency)(svc)

	endpoint := MakeBizEndpoint(svc)

	//rateBucket := ratelimit.NewBucket(time.Second*1, 1)
	//endpoint = NewTokenBucketLimiterWithJuju(rateBucket)(endpoint)

	rateBucket := rate.NewLimiter(rate.Every(time.Second*1), 100)
	endpoint = NewTokenBucketLimiterWithBuildIn(rateBucket)(endpoint)

	healthEndpoint := MakeHealthEndpoint(svc)

	endpoints := BizEndpoints{
		BizEndpoint:    endpoint,
		HealthEndpoint: healthEndpoint,
	}

	r := MakeHttpHandler(ctx, endpoints, logger)

	registrar := Register(*consulHost, *consulPort, *serviceHost, *servicePort, logger)

	go func() {
		fmt.Println("http server start at port:" + *servicePort)
		registrar.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	registrar.Deregister()
	fmt.Printf("service stop:%v\n", error)
}
