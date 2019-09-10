package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"strings"
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

type BizEndpoints struct {
	BizEndpoint    endpoint.Endpoint
	HealthEndpoint endpoint.Endpoint
	AuthEndpoint   endpoint.Endpoint
}

var (
	ErrInvalidType = errors.New("requestType has 4 types: Add, Sub, Mul and Div")
)

func MakeBizEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*BizRequest)
		fmt.Printf("req：#%v\n", req)
		var (
			res, a, b int
			calError  error
		)

		a = req.A
		b = req.B
		bizRes := &BizResponse{}

		if strings.EqualFold(req.ReqType, "Add") {
			res = svc.Add(a, b)
		} else if strings.EqualFold(req.ReqType, "Sub") {
			res = svc.Sub(a, b)
		} else if strings.EqualFold(req.ReqType, "Mul") {
			res = svc.Mul(a, b)
		} else if strings.EqualFold(req.ReqType, "Div") {
			res, calError = svc.Div(a, b)
		} else {
			return bizRes, ErrInvalidType
		}
		bizRes.Result = res
		bizRes.Error = calError
		return bizRes, nil
	}
}

type HealthRequest struct {
}

type HealthResponse struct {
	Status bool `json:"status"`
}

func MakeHealthEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := srv.HealthCheck()
		bizRes := &HealthResponse{}
		bizRes.Status = status
		return bizRes, nil
	}
}

type AuthRequest struct {
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Error   string `json:"error"`
}

func MakeAuthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*AuthRequest)

		token, err := svc.Login(req.Name, req.Pwd)

		resp := &AuthResponse{}
		if err != nil {
			resp.Success = false
			resp.Token = token
			resp.Error = err.Error()
		} else {
			resp.Success = true
			resp.Token = token
		}
		return resp, nil
	}
}
