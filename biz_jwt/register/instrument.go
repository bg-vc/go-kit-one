package main

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/juju/ratelimit"
	"golang.org/x/time/rate"
	"time"
)

var (
	ErrLimitExceed = errors.New("rate limit exceed")
)

func NewTokenBucketLimiterWithJuju(bkt *ratelimit.Bucket) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if bkt.TakeAvailable(1) == 0 {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}

func NewTokenBucketLimiterWithBuildIn(bkt *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bkt.Allow() {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}

type MetricMiddleware struct {
	Service
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
}

func NewMetrics(counter metrics.Counter, histogram metrics.Histogram) ServiceMiddleware {
	return func(next Service) Service {
		return MetricMiddleware{
			Service:        next,
			RequestCount:   counter,
			RequestLatency: histogram,
		}
	}
}

func (mw MetricMiddleware) Add(a, b int) (ret int) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Add"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	ret = mw.Service.Add(a, b)
	return
}

func (mw MetricMiddleware) Sub(a, b int) (ret int) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Sub"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	ret = mw.Service.Sub(a, b)
	return
}

func (mw MetricMiddleware) Mul(a, b int) (ret int) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Mul"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	ret = mw.Service.Mul(a, b)
	return
}

func (mw MetricMiddleware) Div(a, b int) (ret int, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Div"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	ret, err = mw.Service.Div(a, b)
	return
}

func (mw MetricMiddleware) HealthCheck() (ret bool) {
	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	ret = mw.Service.HealthCheck()
	return
}

func (mw MetricMiddleware) Login(name, pwd string) (token string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Login"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	token, err = mw.Service.Login(name, pwd)
	return
}
