package main

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"time"
)

func MakeDiscoverEndpoint(ctx context.Context, client consul.Client, logger log.Logger) endpoint.Endpoint {
	serviceName := "biz"
	tags := []string{"biz", "vc"}
	passingOnly := true
	duration := 500 * time.Millisecond

	instancer := consul.NewInstancer(client, logger, serviceName, tags, passingOnly)

	factory := BizFactory(ctx, "POST", "biz")

	endpointer := sd.NewEndpointer(instancer, factory, logger)

	balancer := lb.NewRoundRobin(endpointer)

	retry := lb.Retry(1, duration, balancer)

	return retry
}
