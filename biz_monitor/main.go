package main

import (
	"context"
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

func main() {
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

	r := MakeHttpHandler(ctx, endpoint, logger)

	go func() {
		fmt.Println("http server start at port:8000")
		handler := r
		errChan <- http.ListenAndServe(":8000", handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("http.ListenAndServe error:%v\n", <-errChan)
}
