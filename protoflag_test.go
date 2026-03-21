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

package protoflag_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/daishe/protoflag"
	"github.com/daishe/protoflag/internal/genreflect"
)

func TestName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given []protoreflect.FieldDescriptor
		want  string
	}{
		{
			name:  "empty-path",
			given: []protoreflect.FieldDescriptor{},
			want:  "input",
		},
		{
			name:  "single-field",
			given: []protoreflect.FieldDescriptor{genreflect.Bottom_Value_field},
			want:  string(genreflect.Bottom_Value_field.Name()),
		},
		{
			name:  "multiple-fields",
			given: []protoreflect.FieldDescriptor{genreflect.Middle_Bottom_field, genreflect.Bottom_Value_field},
			want:  fmt.Sprintf("%s.%s", genreflect.Middle_Bottom_field.Name(), genreflect.Bottom_Value_field.Name()),
		},
		{
			name:  "long-name",
			given: []protoreflect.FieldDescriptor{genreflect.Bottom_ValueWithLongName_field},
			want:  string(genreflect.Bottom_ValueWithLongName_field.Name()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := protoflag.Name(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestTextName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given []protoreflect.FieldDescriptor
		want  string
	}{
		{
			name:  "empty-path",
			given: []protoreflect.FieldDescriptor{},
			want:  "input",
		},
		{
			name:  "single-field",
			given: []protoreflect.FieldDescriptor{genreflect.Bottom_Value_field},
			want:  genreflect.Bottom_Value_field.TextName(),
		},
		{
			name:  "multiple-fields",
			given: []protoreflect.FieldDescriptor{genreflect.Middle_Bottom_field, genreflect.Bottom_Value_field},
			want:  fmt.Sprintf("%s.%s", genreflect.Middle_Bottom_field.TextName(), genreflect.Bottom_Value_field.TextName()),
		},
		{
			name:  "long-name",
			given: []protoreflect.FieldDescriptor{genreflect.Bottom_ValueWithLongName_field},
			want:  genreflect.Bottom_ValueWithLongName_field.TextName(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := protoflag.TextName(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestJSONName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given []protoreflect.FieldDescriptor
		want  string
	}{
		{
			name:  "empty-path",
			given: []protoreflect.FieldDescriptor{},
			want:  "input",
		},
		{
			name:  "single-field",
			given: []protoreflect.FieldDescriptor{genreflect.Bottom_Value_field},
			want:  genreflect.Bottom_Value_field.JSONName(),
		},
		{
			name:  "multiple-fields",
			given: []protoreflect.FieldDescriptor{genreflect.Middle_Bottom_field, genreflect.Bottom_Value_field},
			want:  fmt.Sprintf("%s.%s", genreflect.Middle_Bottom_field.JSONName(), genreflect.Bottom_Value_field.JSONName()),
		},
		{
			name:  "long-name",
			given: []protoreflect.FieldDescriptor{genreflect.Bottom_ValueWithLongName_field},
			want:  genreflect.Bottom_ValueWithLongName_field.JSONName(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := protoflag.JSONName(test.given)
			require.Equal(t, test.want, got)
		})
	}
}
