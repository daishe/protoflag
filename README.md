# protoflag

**protoflag** is a tiny Go library that can turn any [Protocol Buffer](https://protobuf.dev/) message into a set of command‑line flags. It works both with the standard flag package and with the popular [pflag package](https://github.com/spf13/pflag), handling nested fields, messages, lists and maps automatically.

[![Go Reference](https://pkg.go.dev/badge/github.com/daishe/protoflag.svg)](https://pkg.go.dev/github.com/daishe/protoflag)
[![Go Report Card](https://goreportcard.com/badge/github.com/daishe/protoflag)](https://goreportcard.com/report/github.com/daishe/protoflag)

> Looking for a package that works similarly but on any Go type? Check [jsonflag](https://github.com/daishe/jsonflag).

## Adding to project

First, use `go get` to download and add the latest version of the library to the project.

```sh
go get -u github.com/daishe/protoflag
```

Then include in your source code.

```go
import "github.com/daishe/protoflag"
```

## Usage example (with standard go `flag` package)

```go
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
```

## License

The project is released under the **Apache License, Version 2.0**. See the full LICENSE file for the complete terms and conditions.
