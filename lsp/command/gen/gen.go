// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gen is used to generate command bindings from the gopls command
// interface.
package gen

import (
	"bytes"
	"fmt"
	"go/types"
	"text/template"

	"github.com/kevinswiber/languageserver-go/imports"
	"github.com/kevinswiber/languageserver-go/lsp/command/commandmeta"
)

const src = `// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Don't include this file during code generation, or it will break the build
// if existing interface methods have been modified.
//go:build !generate
// +build !generate

package command

// Code generated by generate.go. DO NOT EDIT.

import (
	{{range $k, $v := .Imports -}}
	"{{$k}}"
	{{end}}
)

const (
{{- range .Commands}}
	{{.MethodName}} Command = "{{.Name}}"
{{- end}}
)

var Commands = []Command {
{{- range .Commands}}
	{{.MethodName}},
{{- end}}
}

func Dispatch(ctx context.Context, params *protocol.ExecuteCommandParams, s Interface) (interface{}, error) {
	switch params.Command {
	{{- range .Commands}}
	case "{{.ID}}":
		{{- if .Args -}}
			{{- range $i, $v := .Args}}
		var a{{$i}} {{typeString $v.Type}}
			{{- end}}
		if err := UnmarshalArgs(params.Arguments{{range $i, $v := .Args}}, &a{{$i}}{{end}}); err != nil {
			return nil, err
		}
		{{end -}}
		return {{if not .Result}}nil, {{end}}s.{{.MethodName}}(ctx{{range $i, $v := .Args}}, a{{$i}}{{end}})
	{{- end}}
	}
	return nil, fmt.Errorf("unsupported command %q", params.Command)
}
{{- range .Commands}}

func New{{.MethodName}}Command(title string, {{range $i, $v := .Args}}{{if $i}}, {{end}}a{{$i}} {{typeString $v.Type}}{{end}}) (protocol.Command, error) {
	args, err := MarshalArgs({{range $i, $v := .Args}}{{if $i}}, {{end}}a{{$i}}{{end}})
	if err != nil {
		return protocol.Command{}, err
	}
	return protocol.Command{
		Title: title,
		Command: "{{.ID}}",
		Arguments: args,
	}, nil
}
{{end}}
`

type data struct {
	Imports  map[string]bool
	Commands []*commandmeta.Command
}

func Generate() ([]byte, error) {
	pkg, cmds, err := commandmeta.Load()
	if err != nil {
		return nil, fmt.Errorf("loading command data: %v", err)
	}
	qf := func(p *types.Package) string {
		if p == pkg.Types {
			return ""
		}
		return p.Name()
	}
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"typeString": func(t types.Type) string {
			return types.TypeString(t, qf)
		},
	}).Parse(src)
	if err != nil {
		return nil, err
	}
	d := data{
		Commands: cmds,
		Imports: map[string]bool{
			"context": true,
			"fmt":     true,
			"github.com/kevinswiber/languageserver-go/lsp/protocol": true,
		},
	}
	const thispkg = "github.com/kevinswiber/languageserver-go/lsp/command"
	for _, c := range d.Commands {
		for _, arg := range c.Args {
			pth := pkgPath(arg.Type)
			if pth != "" && pth != thispkg {
				d.Imports[pth] = true
			}
		}
		pth := pkgPath(c.Result)
		if pth != "" && pth != thispkg {
			d.Imports[pth] = true
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return nil, fmt.Errorf("executing: %v", err)
	}

	opts := &imports.Options{
		AllErrors:  true,
		FormatOnly: true,
		Comments:   true,
	}
	content, err := imports.Process("", buf.Bytes(), opts)
	if err != nil {
		return nil, fmt.Errorf("goimports: %v", err)
	}
	return content, nil
}

func pkgPath(t types.Type) string {
	if n, ok := t.(*types.Named); ok {
		if pkg := n.Obj().Pkg(); pkg != nil {
			return pkg.Path()
		}
	}
	return ""
}
