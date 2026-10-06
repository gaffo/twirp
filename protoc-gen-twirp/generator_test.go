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

package main

import (
	"os"
	"os/exec"
	"testing"

	"google.golang.org/protobuf/proto"
	plugin "google.golang.org/protobuf/types/pluginpb"
)

func TestGenerateParseCommandLineParamsError(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		g := &twirp{}
		g.Generate(&plugin.CodeGeneratorRequest{
			Parameter: proto.String("invalid"),
		})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateParseCommandLineParamsError")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestDeprecationParagraph(t *testing.T) {
	tests := []struct {
		name    string
		leading string
		want    []string
	}{
		{"empty", "", nil},
		{"no notice", " Foo does things.\n", nil},
		{"mid-sentence mention", " Foo is not\n Deprecated: really.\n", nil},
		{"only notice", " Deprecated: Use Bar.\n", []string{"Deprecated: Use Bar."}},
		{
			"notice paragraph",
			" Foo does things.\n\n Deprecated: Use Bar\n instead.\n\n More text.\n",
			[]string{"Deprecated: Use Bar", "instead."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deprecationParagraph(tt.leading)
			if len(got) != len(tt.want) {
				t.Fatalf("deprecationParagraph(%q) = %q, want %q", tt.leading, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("deprecationParagraph(%q) = %q, want %q", tt.leading, got, tt.want)
				}
			}
		})
	}
}
