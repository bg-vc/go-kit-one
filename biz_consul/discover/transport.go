package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
)

type BizRequest struct {
	ReqType string `json:"type"`
	A       int    `json:"a"`
	B       int    `json:"b"`
}

type BizResponse struct {
	Result int   `json:"result"`
	Error  error `json:"error"`
}

func MakeHttpHandler(endpoint endpoint.Endpoint, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorLogger(logger),
		kitHttp.ServerErrorEncoder(kitHttp.DefaultErrorEncoder),
	}

	r.Methods("POST").Path("/biz").Handler(kitHttp.NewServer(
		endpoint,
		decodeDiscoverRequest,
		encodeDiscoverResponse,
		options...
	))

	return r
}

func decodeDiscoverRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request BizRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	fmt.Printf("request:#%v\n", request)
	return request, nil
}

func encodeDiscoverResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
