package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	kitHttp "github.com/go-kit/kit/transport/http"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func BizFactory(ctx context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint endpoint.Endpoint, closer io.Closer, err error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path
		var (
			enc kitHttp.EncodeRequestFunc
			dec kitHttp.DecodeResponseFunc
		)
		enc, dec = encodeBizRequest, decodeVizResponse

		return kitHttp.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil
	}
}

func encodeBizRequest(ctx context.Context, r *http.Request, request interface{}) error {
	bizReq := request.(BizRequest)
	p := "/" + bizReq.ReqType + "/" + strconv.Itoa(bizReq.A) + "/" + strconv.Itoa(bizReq.B)
	r.URL.Path += p
	fmt.Printf("r.URL.Path:%v\n", r.URL.Path)
	return nil
}

func decodeVizResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	response := &BizResponse{}
	var s map[string]interface{}

	if respCode := resp.StatusCode; respCode >= 400 {
		if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
			return nil, err
		}
		return nil, errors.New(s["error"].(string) + "\n")
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}

	return response, nil
}
