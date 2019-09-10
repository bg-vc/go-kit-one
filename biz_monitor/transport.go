package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
)

var (
	ErrBadRequest = errors.New("invalid request parameter")
)

func MakeHttpHandler(ctx context.Context, endpoint endpoint.Endpoint, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorLogger(logger),
		kitHttp.ServerErrorEncoder(kitHttp.DefaultErrorEncoder),
	}

	r.Methods("POST").Path("/biz/{type}/{a}/{b}").Handler(kitHttp.NewServer(
		endpoint,
		decodeBizRequest,
		encodeBizResponse,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	return r
}

func decodeBizRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	fmt.Printf("vars:%v\n", vars)

	reqType, ok := vars["type"]
	if !ok {
		return nil, ErrBadRequest
	}
	pa, ok := vars["a"]
	if !ok {
		return nil, ErrBadRequest
	}
	pb, ok := vars["b"]
	if !ok {
		return nil, ErrBadRequest
	}
	a, _ := strconv.Atoi(pa)
	b, _ := strconv.Atoi(pb)

	return &BizRequest{
		ReqType: reqType,
		A:       a,
		B:       b,
	}, nil
}

func encodeBizResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fmt.Printf("res:%#v\n", response)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
