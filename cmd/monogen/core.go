package main

var monolith = "github.com/orangootan/monolith/pkg/monolith"

type File struct {
	Package
	Types      []Type
	Methods    []Method
	Interfaces []Interface
}

type Decl struct {
	Comments []string
	Name     string
}

type Package struct {
	Decl
}

type Type struct {
	Decl
}

type Interface struct {
	Decl
	Methods []Function
}

type Method struct {
	Receiver *ValueGroup
	Function
}

type Function struct {
	Decl
	Params  []ValueGroup
	Results []ValueGroup
}

type ValueGroup struct {
	Names []string
	Type  string
}

type Service struct {
	Type        Type
	Methods     []Method
	Constructor *Method
}

type ServiceMap map[[2]string]*Service
