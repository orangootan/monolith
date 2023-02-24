package main

import (
	"go/ast"
	"go/token"
)

func valueGroupFromField(field *ast.Field) ValueGroup {
	var vg ValueGroup
	vg.Type = field.Type.(*ast.Ident).Name
	for _, name := range field.Names {
		vg.Names = append(vg.Names, name.Name)
	}
	return vg
}

func parseInterfaceMethod(field *ast.Field) Function {
	var f Function
	f.Name = field.Names[0].Name
	if doc := field.Doc; doc != nil {
		for _, comment := range doc.List {
			f.Comments = append(f.Comments, comment.Text)
		}
	}
	funcType := field.Type.(*ast.FuncType)
	f.Params, f.Results = parseFuncType(funcType)
	return f
}

func parseGenDecl(d *ast.GenDecl) (*Type, *Interface) {
	if d.Tok != token.TYPE {
		return nil, nil
	}
	var decl Decl
	if d.Doc != nil {
		for _, comment := range d.Doc.List {
			decl.Comments = append(decl.Comments, comment.Text)
		}
	}
	typeSpec := d.Specs[0].(*ast.TypeSpec)
	decl.Name = typeSpec.Name.Name
	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return &Type{
			Decl: decl,
		}, nil
	}
	var methods []Function
	if ims := interfaceType.Methods; ims != nil {
		for _, m := range ims.List {
			methods = append(methods, parseInterfaceMethod(m))
		}
	}
	return nil, &Interface{
		Decl:    decl,
		Methods: methods,
	}
}

func parseFile(f *ast.File) File {
	var file File
	file.Name = f.Name.Name
	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			fd := decl.(*ast.FuncDecl)
			method := parseFuncDecl(fd)
			file.Methods = append(file.Methods, method)
		case *ast.GenDecl:
			gd := decl.(*ast.GenDecl)
			t, i := parseGenDecl(gd)
			if t != nil {
				file.Types = append(file.Types, *t)
			}
			if i != nil {
				file.Interfaces = append(file.Interfaces, *i)
			}
		}
	}
	return file
}

func parseFuncDecl(d *ast.FuncDecl) Method {
	var receiver *ValueGroup
	if d.Recv != nil && len(d.Recv.List) > 0 {
		var r ValueGroup
		field := d.Recv.List[0]
		r.Type = field.Type.(*ast.Ident).Name
		for _, name := range field.Names {
			r.Names = append(r.Names, name.Name)
		}
		receiver = &r
	}
	var f Function
	f.Name = d.Name.Name
	if d.Doc != nil {
		for _, comment := range d.Doc.List {
			f.Comments = append(f.Comments, comment.Text)
		}
	}
	f.Params, f.Results = parseFuncType(d.Type)
	return Method{
		Receiver: receiver,
		Function: f,
	}
}

func parseFuncType(f *ast.FuncType) ([]ValueGroup, []ValueGroup) {
	var params, results []ValueGroup
	for _, param := range f.Params.List {
		vg := valueGroupFromField(param)
		params = append(params, vg)
	}
	for _, result := range f.Results.List {
		vg := valueGroupFromField(result)
		results = append(results, vg)
	}
	return params, results
}

func (d Decl) isIgnored() bool {
	return find("//monolith:ignore", d.Comments) > -1
}

func (d Decl) isService() bool {
	return find("//monolith:service", d.Comments) > -1
}

func (f File) filter() File {
	var ts []Type
	for _, t := range f.Types {
		if t.isService() && !t.isIgnored() {
			ts = append(ts, t)
		}
	}
	var is []Interface
	for _, i := range f.Interfaces {
		if i.isService() && !i.isIgnored() {
			var ms []Function
			for _, m := range i.Methods {
				if !m.isIgnored() {
					ms = append(ms, m)
				}
			}
			i.Methods = ms
			is = append(is, i)
		}
	}
	var ms []Method
	for _, m := range f.Methods {
		if !m.isIgnored() {
			ms = append(ms, m)
		}
	}
	return File{
		Package:    f.Package,
		Types:      ts,
		Methods:    ms,
		Interfaces: is,
	}
}

func createServiceMap(fs []File) ServiceMap {
	services := make(map[[2]string]*Service)
	for _, f := range fs {
		for _, t := range f.Types {
			services[[2]string{f.Package.Name, t.Name}] = &Service{
				Type: t,
			}
		}
	}
	for _, f := range fs {
		for _, m := range f.Methods {
			r := m.Receiver
			if r != nil {
				service, ok := services[[2]string{f.Package.Name, r.Type}]
				if ok {
					service.Methods = append(service.Methods, m)
				}
			}
			isConstructor := r == nil &&
				len(m.Params) == 1 &&
				len(m.Params[0].Names) == 1 &&
				m.Params[0].Type == "string" &&
				len(m.Results) == 2 &&
				len(m.Results[0].Names) == 1 &&
				len(m.Results[1].Names) == 1 &&
				m.Results[1].Type == "error"
			if isConstructor {
				service, ok := services[[2]string{f.Package.Name, m.Results[0].Type}]
				if !ok || service.Constructor != nil {
					continue
				}
				service.Constructor = &m
			}
		}
	}
	return services
}
