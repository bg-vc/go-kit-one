package main

import (
	"errors"
)

type Service interface {
	Add(a, b int) int

	Sub(a, b int) int

	Mul(a, b int) int

	Div(a, b int) (int, error)

	HealthCheck() bool
}

type BizService struct {
}

func NewBizService() Service {
	return &BizService{}
}

func (s *BizService) Add(a, b int) int {
	return a + b
}

func (s *BizService) Sub(a, b int) int {
	return a - b
}

func (s *BizService) Mul(a, b int) int {
	return a * b
}

func (s *BizService) Div(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("the divisor is zero")
	}
	return a / b, nil
}

func (s *BizService) HealthCheck() bool {
	return true
}

type ServiceMiddleware func(Service) Service
