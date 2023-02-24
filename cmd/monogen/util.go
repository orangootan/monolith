package main

import "strconv"

func find[T comparable](item T, items []T) int {
	for i, it := range items {
		if it == item {
			return i
		}
	}
	return -1
}

func Map[T any, S any](s []T, fn func(T) S) []S {
	var result []S
	for _, item := range s {
		result = append(result, fn(item))
	}
	return result
}

type nameSelector struct {
	names map[string]bool
}

func newNameSelector() nameSelector {
	return nameSelector{
		names: make(map[string]bool),
	}
}

func (ns *nameSelector) Add(name string) {
	ns.names[name] = true
}

func (ns *nameSelector) New(base string) string {
	i := 1
	name := base
	for ns.names[name] {
		i++
		name = base + strconv.Itoa(i)
	}
	ns.Add(name)
	return name
}
