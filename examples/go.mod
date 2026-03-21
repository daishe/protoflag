module github.com/daishe/protoflag/examples

go 1.26.1

replace github.com/daishe/protoflag => ../

require (
	github.com/daishe/protoflag v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/daishe/jsonflag v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10
)
