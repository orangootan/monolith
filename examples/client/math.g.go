package main

import m "github.com/orangootan/monolith/pkg/monolith"

type MathProxy m.Instance

func init() {
	m.RegisterProxy[Math](func(i m.Instance) any {
		return MathProxy(i)
	})
}
func (p MathProxy) Add(a, b int) (c int, err error) {
	params := struct {
		A, B int
	}{
		A: a,
		B: b,
	}
	var results struct {
		C   int
		Err error
	}
	err2 := m.Instance(p).Call("Add", params, &results)
	if err2 != nil {
		results.Err = err2
	}
	return results.C, results.Err
}
func (p MathProxy) Divide(a int, b int) (int, error) {
	params := struct {
		A int
		B int
	}{
		A: a,
		B: b,
	}
	var results struct {
		R  int
		R2 error
	}
	err := m.Instance(p).Call("Divide", params, &results)
	if err != nil {
		results.R2 = err
	}
	return results.R, results.R2
}
func (p MathProxy) Sqrt(x float64) float64 {
	params := struct {
		X float64
	}{X: x}
	var results struct {
		R float64
	}
	err := m.Instance(p).Call("Sqrt", params, &results)
	if err != nil {
		panic(err)
	}
	return results.R
}
