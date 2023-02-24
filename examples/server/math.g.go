package main

import m "github.com/orangootan/monolith/pkg/monolith"

func init() {
	m.RegisterTypeHandler("Math", MathHandler)
}
func MathHandler(id string, method string, decode func(params any) error, encode func(params any) error) (err error) {
	instance, err := MathFromString(id)
	if err != nil {
		return
	}
	switch method {
	case "Add":
		var params struct {
			A, B int
		}
		var results struct {
			C   int
			Err error
		}
		err = decode(&params)
		if err != nil {
			return
		}
		results.C, results.Err = instance.Add(params.A, params.B)
		return encode(results)
	case "Divide":
		var params struct {
			A int
			B int
		}
		var results struct {
			R  int
			R2 error
		}
		err = decode(&params)
		if err != nil {
			return
		}
		results.R, results.R2 = instance.Divide(params.A, params.B)
		return encode(results)
	case "Sqrt":
		var params struct {
			X float64
		}
		var results struct {
			R float64
		}
		err = decode(&params)
		if err != nil {
			return
		}
		results.R = instance.Sqrt(params.X)
		return encode(results)
	default:
		return m.MethodNotFoundError
	}
}
