package main

import (
	j "github.com/dave/jennifer/jen"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var toTitle = cases.Title(language.English)

func generateProxyMethod(proxy string, method Function) j.Code {
	nsGlobal := newNameSelector()
	nsParams := newNameSelector()
	nsResults := newNameSelector()
	var paramNames []string
	var paramNamesTitle []string
	var resultNamesTitle []string
	paramGroups := Map(method.Params, func(vg ValueGroup) j.Code {
		names := Map(vg.Names, func(name string) j.Code {
			paramNames = append(paramNames, name)
			nsGlobal.Add(name)
			return j.Id(name)
		})
		return j.List(names...).Id(vg.Type)
	})
	paramGroupsTitle := Map(method.Params, func(vg ValueGroup) j.Code {
		names := Map(vg.Names, func(name string) j.Code {
			title := nsParams.New(toTitle.String(name))
			paramNamesTitle = append(paramNamesTitle, title)
			return j.Id(title)
		})
		return j.List(names...).Id(vg.Type)
	})
	resultGroups := Map(method.Results, func(vg ValueGroup) j.Code {
		names := Map(vg.Names, func(name string) j.Code {
			nsGlobal.Add(name)
			return j.Id(name)
		})
		return j.List(names...).Id(vg.Type)
	})
	resultGroupsTitle := Map(method.Results, func(vg ValueGroup) j.Code {
		names := Map(vg.Names, func(name string) j.Code {
			title := nsResults.New(toTitle.String(name))
			resultNamesTitle = append(resultNamesTitle, title)
			return j.Id(title)
		})
		if len(names) == 0 {
			name := nsResults.New("R")
			names = append(names, j.Id(name))
			resultNamesTitle = append(resultNamesTitle, name)
		}
		return j.List(names...).Id(vg.Type)
	})
	results := Map(resultNamesTitle, func(name string) j.Code {
		return j.Id("results").Dot(name)
	})
	lastIsError := method.Results[len(method.Results)-1].Type == "error"
	errorName := nsGlobal.New("err")
	ifError := func(g *j.Group) {
		g.If(j.Id(errorName).Op("!=").Nil()).BlockFunc(func(g1 *j.Group) {
			if lastIsError {
				g1.Id("results").Dot(resultNamesTitle[len(resultNamesTitle)-1]).Op("=").Id(errorName)
			} else {
				g1.Panic(j.Id(errorName))
			}
		})
	}
	return j.Func().Params(j.Id("p").Id(proxy + "Proxy")).Id(method.Name).Params(paramGroups...).Params(resultGroups...).BlockFunc(func(g *j.Group) {
		g.Id("params").Op(":=").Struct(paramGroupsTitle...).Values(j.DictFunc(func(d j.Dict) {
			for i := 0; i < len(paramNames); i++ {
				d[j.Id(paramNamesTitle[i])] = j.Id(paramNames[i])
			}
		}))
		g.Var().Id("results").Struct(resultGroupsTitle...)
		g.Id(errorName).Op(":=").Qual(monolith, "Instance").Call(j.Id("p")).Dot("Call").Call(
			j.Lit(method.Name), j.Id("params"), j.Op("&").Id("results"))
		ifError(g)
		g.Return(results...)
	})
}

func generateRegisterProxy(name string) j.Code {
	return j.Qual(monolith, "RegisterProxy").Types(j.Id(name)).
		Call(j.Func().Params(j.Id("i").Qual(monolith, "Instance")).Any().Block(
			j.Return(j.Id(name + "Proxy").Call(j.Id("i")))))
}

func generateTypeHandler(s *Service) j.Code {
	return j.Func().Id(s.Type.Name+"Handler").Params(
		j.Id("id").Id("string"),
		j.Id("method").Id("string"),
		j.Id("decode").Func().Params(j.Id("params").Id("any")).Params(j.Id("error")),
		j.Id("encode").Func().Params(j.Id("params").Id("any")).Params(j.Id("error")),
	).Params(j.Id("err").Id("error")).BlockFunc(func(g *j.Group) {
		g.List(j.Id("instance"), j.Id("err")).Op(":=").Id(s.Constructor.Name).Call(j.Id("id"))
		g.If(j.Id("err").Op("!=").Nil()).Block(j.Return())
		g.Switch(j.Id("method")).BlockFunc(func(g1 *j.Group) {
			for _, m := range s.Methods {
				nsParams := newNameSelector()
				nsResults := newNameSelector()
				var paramNamesTitle []string
				var resultNamesTitle []string
				paramGroupsTitle := Map(m.Params, func(vg ValueGroup) j.Code {
					names := Map(vg.Names, func(name string) j.Code {
						title := nsParams.New(toTitle.String(name))
						paramNamesTitle = append(paramNamesTitle, title)
						return j.Id(title)
					})
					return j.List(names...).Id(vg.Type)
				})
				resultGroupsTitle := Map(m.Results, func(vg ValueGroup) j.Code {
					names := Map(vg.Names, func(name string) j.Code {
						title := nsResults.New(toTitle.String(name))
						resultNamesTitle = append(resultNamesTitle, title)
						return j.Id(title)
					})
					if len(names) == 0 {
						name := nsResults.New("R")
						names = append(names, j.Id(name))
						resultNamesTitle = append(resultNamesTitle, name)
					}
					return j.List(names...).Id(vg.Type)
				})
				params := Map(paramNamesTitle, func(name string) j.Code {
					return j.Id("params").Dot(name)
				})
				results := Map(resultNamesTitle, func(name string) j.Code {
					return j.Id("results").Dot(name)
				})
				g1.Case(j.Lit(m.Name)).BlockFunc(func(g2 *j.Group) {
					g2.Var().Id("params").Struct(paramGroupsTitle...)
					g2.Var().Id("results").Struct(resultGroupsTitle...)
					g2.Id("err").Op("=").Id("decode").Call(j.Op("&").Id("params"))
					g2.If(j.Id("err").Op("!=").Nil()).Block(j.Return())
					g2.List(results...).Op("=").Id("instance").Dot(m.Name).Call(params...)
					g2.Return(j.Id("encode").Call(j.Id("results")))
				})
			}
			g1.Default().Block(j.Return(j.Qual(monolith, "MethodNotFoundError")))
		})
	})
}

func generateFile(s File, sm ServiceMap) *j.File {
	f := j.NewFile(s.Package.Name)
	f.ImportAlias(monolith, "m")
	for _, i := range s.Interfaces {
		f.Type().Id(i.Name+"Proxy").Qual(monolith, "Instance")
	}
	if len(s.Interfaces) != 0 {
		f.Func().Id("init").Params().BlockFunc(func(g *j.Group) {
			for _, i := range s.Interfaces {
				g.Add(generateRegisterProxy(i.Name))
			}
		})
	}
	for _, i := range s.Interfaces {
		for _, m := range i.Methods {
			f.Add(generateProxyMethod(i.Name, m))
		}
	}
	if len(s.Types) != 0 {
		f.Func().Id("init").Params().BlockFunc(func(g *j.Group) {
			for _, t := range s.Types {
				g.Qual(monolith, "RegisterTypeHandler").Call(j.Lit(t.Name), j.Id(t.Name+"Handler"))
			}
		})
	}
	for _, t := range s.Types {
		service := sm[[2]string{s.Package.Name, t.Name}]
		f.Add(generateTypeHandler(service))
	}
	return f
}
