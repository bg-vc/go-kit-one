package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	kitJwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	goZipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
)

var (
	ErrBadRequest = errors.New("invalid request parameter")
)

func MakeHttpHandler(ctx context.Context, endpoints BizEndpoints, tracer *goZipkin.Tracer, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	zipkinServer := zipkin.HTTPServerTrace(tracer, zipkin.Name("http-transport"))

	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorLogger(logger),
		kitHttp.ServerErrorEncoder(kitHttp.DefaultErrorEncoder),
		zipkinServer,
	}

	r.Methods("POST").Path("/biz/{type}/{a}/{b}").Handler(kitHttp.NewServer(
		endpoints.BizEndpoint,
		decodeBizRequest,
		encodeBizResponse,
		append(options, kitHttp.ServerBefore(kitJwt.HTTPToContext()))...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	r.Methods("GET").Path("/health").Handler(kitHttp.NewServer(
		endpoints.HealthEndpoint,
		decodeHealthRequest,
		encodeBizResponse,
		options...,
	))

	r.Methods("POST").Path("/login").Handler(kitHttp.NewServer(
		endpoints.AuthEndpoint,
		decodeLoginRequest,
		encodeLoginResponse,
		options...,
	))

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

func decodeHealthRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return &HealthRequest{}, nil
}

func encodeLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func decodeLoginRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	loginRequest := &AuthRequest{}
	fmt.Printf("r.Body:#%v\n", r.Body)

	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		return nil, err
	}
	fmt.Printf("loginRequest:#%v\n", loginRequest)
	return loginRequest, nil
}
