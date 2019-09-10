package main

import (
	"github.com/go-kit/kit/log"
	"time"
)

type LoggingMiddleware struct {
	Service
	logger log.Logger
}

func NewLoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return LoggingMiddleware{next, logger}
	}
}

func (mw LoggingMiddleware) Add(a, b int) (ret int) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Add",
			"a", a,
			"b", b,
			"result", ret,
			"cost", time.Since(begin),
		)

	}(time.Now())

	ret = mw.Service.Add(a, b)
	return
}

func (mw LoggingMiddleware) Sub(a, b int) (ret int) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Sub",
			"a", a,
			"b", b,
			"result", ret,
			"cost", time.Since(begin),
		)

	}(time.Now())

	ret = mw.Service.Sub(a, b)
	return
}

func (mw LoggingMiddleware) Mul(a, b int) (ret int) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Mul",
			"a", a,
			"b", b,
			"result", ret,
			"cost", time.Since(begin),
		)

	}(time.Now())

	ret = mw.Service.Mul(a, b)
	return
}

func (mw LoggingMiddleware) Div(a, b int) (ret int, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Div",
			"a", a,
			"b", b,
			"result", ret,
			"cost", time.Since(begin),
		)

	}(time.Now())

	ret, err = mw.Service.Div(a, b)
	return
}

func (mw LoggingMiddleware) HealthCheck() (ret bool) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "HealthCheck",
			"result", ret,
			"cost", time.Since(begin),
		)
	}(time.Now())

	ret = mw.Service.HealthCheck()
	return
}
