// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package deprecated

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// deprecationNotice returns the "Deprecated:" paragraph of a doc comment, or
// "" if there is none.
func deprecationNotice(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	for _, para := range strings.Split(doc.Text(), "\n\n") {
		if strings.HasPrefix(para, "Deprecated:") {
			return strings.TrimSpace(para)
		}
	}
	return ""
}

func TestDeprecatedDocComments(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "deprecated.twirp.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing generated file: %v", err)
	}

	got := map[string]string{} // identifier -> deprecation notice
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			if d.Recv != nil {
				recv := d.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
				name = recv + "." + name
			}
			got[name] = deprecationNotice(d.Doc)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				got[ts.Name.Name] = deprecationNotice(d.Doc)
				iface, ok := ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				for _, m := range iface.Methods.List {
					for _, n := range m.Names {
						got[ts.Name.Name+"."+n.Name] = deprecationNotice(m.Doc)
					}
				}
			}
		}
	}

	const doNotUse = "Deprecated: Do not use."
	want := map[string]string{
		// Deprecated service.
		"OldService":                  doNotUse,
		"NewOldServiceProtobufClient": doNotUse,
		"NewOldServiceJSONClient":     doNotUse,
		"NewOldServiceServer":         doNotUse,
		"OldService.Method":           "",

		// Service with deprecated methods.
		"NewService":                  "",
		"NewNewServiceProtobufClient": "",
		"NewNewServiceJSONClient":     "",
		"NewNewServiceServer":         "",
		"NewService.Old":              doNotUse,
		"NewService.OldNoComment":     doNotUse,
		"NewService.OldWithNotice":    "Deprecated: Use Current instead.",
		"NewService.Current":          "",
	}
	for _, client := range []string{"newServiceProtobufClient", "newServiceJSONClient"} {
		want[client+".Old"] = doNotUse
		want[client+".OldNoComment"] = doNotUse
		want[client+".OldWithNotice"] = "Deprecated: Use Current instead."
		want[client+".Current"] = ""
	}

	for name, wantNotice := range want {
		gotNotice, ok := got[name]
		if !ok {
			t.Errorf("%s: not found in generated code", name)
			continue
		}
		if gotNotice != wantNotice {
			t.Errorf("%s: deprecation notice = %q, want %q", name, gotNotice, wantNotice)
		}
	}
}

func TestDeprecatedNoticeNotDuplicated(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "deprecated.twirp.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing generated file: %v", err)
	}
	for _, cg := range f.Comments {
		if n := strings.Count(cg.Text(), "Deprecated:"); n > 1 {
			t.Errorf("comment group has %d Deprecated: paragraphs:\n%s", n, cg.Text())
		}
	}
}
