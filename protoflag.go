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

// Package protoflag is a tiny Go library that can turn any Protocol Buffer message into a set of command‑line flags.
package protoflag

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const baseMessageFlagName = "input"

// Name creates a new flag name by joining all names of fields along the provided path.
func Name(path []protoreflect.FieldDescriptor) string {
	b := strings.Builder{}
	for i, p := range path {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(string(p.Name()))
	}
	n := b.String()
	if n == "" {
		return baseMessageFlagName
	}
	return n
}

// TextName creates a new flag name by joining all text names of fields along the provided path.
func TextName(path []protoreflect.FieldDescriptor) string {
	b := strings.Builder{}
	for i, p := range path {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(p.TextName())
	}
	n := b.String()
	if n == "" {
		return baseMessageFlagName
	}
	return n
}

// JSONName creates a new flag name by joining all JSON names of fields along the provided path.
func JSONName(path []protoreflect.FieldDescriptor) string {
	b := strings.Builder{}
	for i, p := range path {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(p.JSONName())
	}
	n := b.String()
	if n == "" {
		return baseMessageFlagName
	}
	return n
}
