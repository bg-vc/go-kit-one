package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	kitPrometheus "github.com/go-kit/kit/metrics/prometheus"
	kitZipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go"
	zipkinHttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	consulHost  = flag.String("consul.host", "192.168.0.103", "consul ip address")
	consulPort  = flag.String("consul.port", "8500", "consul port")
	serviceHost = flag.String("service.host", "192.168.0.103", "service ip address")
	servicePort = flag.String("service.port", "8000", "service port")
	zipkinUrl   = flag.String("zipkin.url", "http://192.168.0.103:9411/api/v2/spans", "zipkin server url")
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

	var zipKinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = *serviceHost + ":" + *servicePort
			serviceName   = "biz-service"
			useNoopTracer = *zipkinUrl == ""
			reporter      = zipkinHttp.NewReporter(*zipkinUrl)
		)

		defer reporter.Close()

		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipKinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("error", err)
			os.Exit(1)
		}
		if !useNoopTracer {
			logger.Log("tracer", "zipkin", "type", "native", "url", *zipkinUrl)
		}
	}

	svc := NewBizService()

	svc = NewLoggingMiddleware(logger)(svc)

	svc = NewMetrics(requestCount, requestLatency)(svc)

	rateBucket := rate.NewLimiter(rate.Every(time.Second*1), 100)

	endpoint := MakeBizEndpoint(svc)

	//rateBucket := ratelimit.NewBucket(time.Second*1, 1)
	//endpoint = NewTokenBucketLimiterWithJuju(rateBucket)(endpoint)

	endpoint = NewTokenBucketLimiterWithBuildIn(rateBucket)(endpoint)
	endpoint = kitZipkin.TraceEndpoint(zipKinTracer, "biz-endpoint")(endpoint)

	healthEndpoint := MakeHealthEndpoint(svc)
	healthEndpoint = NewTokenBucketLimiterWithBuildIn(rateBucket)(healthEndpoint)
	healthEndpoint = kitZipkin.TraceEndpoint(zipKinTracer, "health-endpoint")(healthEndpoint)

	endpoints := BizEndpoints{
		BizEndpoint:    endpoint,
		HealthEndpoint: healthEndpoint,
	}

	r := MakeHttpHandler(ctx, endpoints, zipKinTracer, logger)

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
