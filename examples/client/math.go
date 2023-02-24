package main

//go:generate monogen math.go

//monolith:service
type Math interface {
	Add(a, b int) (c int, err error)
	Divide(a int, b int) (int, error)
	Sqrt(x float64) float64
}
