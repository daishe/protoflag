// Copyright 2026 Marek Dalewski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/daishe/protoflag"
	protoflagtestv1 "github.com/daishe/protoflag/internal/protoflagtest/v1"
)

func main() {
	// message SimpleMessage {
	//   string foo = 1;
	//   string bar = 2;
	//   string baz = 3;
	// }

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	i := &protoflagtestv1.SimpleMessage{}
	for _, val := range protoflag.Recursive(i) {
		pf := &pflag.Flag{
			Name:     protoflag.JSONName(val.Path()),
			Usage:    fmt.Sprintf("usage of --%s flag", protoflag.Name(val.Path())),
			Value:    val,
			DefValue: val.String(),
		}
		if val.IsBoolFlag() {
			pf.NoOptDefVal = "true"
		}
		fs.AddFlag(pf)
	}
	if err := fs.Parse([]string{"--foo=foo value", "--baz=baz value", "--baz=another baz value"}); err != nil {
		panic(err)
	}

	b, err := protojson.MarshalOptions{Indent: "  ", EmitDefaultValues: true}.Marshal(i)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Output:
	// {
	//   "foo": "foo value",
	//   "bar": "",
	//   "baz": "another baz value"
	// }
}
