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
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/daishe/protoflag"
	typesv1 "github.com/daishe/protoflag/examples/simplecurl/types/v1"
)

func main() {
	// config with default values
	config := typesv1.Config_builder{
		Url: typesv1.Url_builder{
			Scheme: "http",
			Host:   "localhost",
			Port:   8080,
		}.Build(),
	}.Build()

	usage := func(path []protoreflect.FieldDescriptor) string {
		opts, ok := path[len(path)-1].Options().(*descriptorpb.FieldOptions)
		if !ok || opts == nil {
			return ""
		}
		usage, ok := proto.GetExtension(opts, typesv1.E_Usage).(string)
		if !ok {
			return ""
		}
		return usage
	}

	fs := flag.NewFlagSet("", flag.ExitOnError) // you can also add it directly to the root flag set
	for _, val := range protoflag.Recursive(config) {
		fs.Var(val, protoflag.JSONName(val.Path()), usage(val.Path()))
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Arguments parsing error: %v\n", err)
		os.Exit(1)
	}

	u := url.URL{
		Scheme: config.GetUrl().GetScheme(),
		Host:   net.JoinHostPort(config.GetUrl().GetHost(), strconv.FormatInt(int64(config.GetUrl().GetPort()), 10)),
		Path:   config.GetUrl().GetPath(),
	}

	resp, err := http.Get(u.String()) //nolint:noctx // this is just an example
	if err != nil {
		fmt.Fprintf(os.Stderr, "Doing request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if config.GetVerbose() {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Dumping response error: %v\n", err)
			os.Exit(1) //nolint:gocritic // this is just an example
		}
		fmt.Printf("%s\n", respBytes)
	}
}
