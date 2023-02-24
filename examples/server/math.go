package main

//go:generate monogen math.go

import (
	"github.com/orangootan/monolith/pkg/monolith"
	"math"
	"strconv"
)

//monolith:service
type Math struct {
	c int
}

func (m Math) Add(a, b int) (c int, err error) {
	c = a + b + m.c
	return
}

func (m Math) Divide(a int, b int) (int, error) {
	if b == 0 {
		return 0, monolith.NewError("divide by zero")
	}
	return a / b, nil
}

func (m Math) Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

func MathFromString(id string) (math Math, err error) {
	math.c, err = strconv.Atoi(id)
	if err != nil {
		err = monolith.NewError(err.Error())
	}
	return
}
