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
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/daishe/protoflag"
	protoflagtestv1 "github.com/daishe/protoflag/internal/protoflagtest/v1"
)

type FlagValueData struct {
	PathFullNames        []string
	Kind                 protoreflect.Kind
	HasMessageDescriptor bool
	HasMessageType       bool
	HasEnumDescriptor    bool
	Type                 string
	Get                  any
	String               string
	IsList               bool
	IsMap                bool
	IsBoolFlag           bool
}

func (d *FlagValueData) Path() string {
	return strings.Join(d.PathFullNames, "/")
}

func DataOfFlagValue(val *protoflag.Value) *FlagValueData {
	pathNames := make([]string, 0, len(val.Path()))
	for _, x := range val.Path() {
		pathNames = append(pathNames, string(x.FullName()))
	}
	return &FlagValueData{
		PathFullNames:        pathNames,
		Kind:                 val.Kind(),
		HasMessageDescriptor: val.MessageDescriptor() != nil,
		HasMessageType:       val.MessageType() != nil,
		HasEnumDescriptor:    val.EnumDescriptor() != nil,
		Type:                 val.Type(),
		Get:                  clone(val.Get()),
		String:               val.String(),
		IsList:               val.IsList(),
		IsMap:                val.IsMap(),
		IsBoolFlag:           val.IsBoolFlag(),
	}
}

func DataOfFlagValues(vals []*protoflag.Value) (d []*FlagValueData) {
	for _, f := range vals {
		d = append(d, DataOfFlagValue(f))
	}
	return d
}

func RequireDataOfFlagValueEqual(t *testing.T, want, got *FlagValueData) {
	t.Helper()
	require.Equal(t, want.PathFullNames, got.PathFullNames, "mismatched Value.Path() result")
	require.Equal(t, want.Kind, got.Kind, "mismatched Value.Kind() result for path %q", want.Path())
	require.Equal(t, want.HasMessageDescriptor, got.HasMessageDescriptor, "mismatched Value.MessageDescriptor() result for path %q", want.Path())
	require.Equal(t, want.HasMessageType, got.HasMessageType, "mismatched Value.MessageType() result for path %q", want.Path())
	require.Equal(t, want.HasEnumDescriptor, got.HasEnumDescriptor, "mismatched Value.EnumDescriptor() result for path %q", want.Path())
	require.Equal(t, want.Type, got.Type, "mismatched Value.Type() result for path %q", want.Path())
	RequireDiffEqual(t, want.Get, got.Get, "mismatched Value.Get() result for path %q (-want, +got)", want.Path())
	require.Equal(t, want.String, got.String, "mismatched Value.String() result for path %q", want.Path())
	require.Equal(t, want.IsList, got.IsList, "mismatched Value.IsList() result for path %q", want.Path())
	require.Equal(t, want.IsMap, got.IsMap, "mismatched Value.IsMap() result for path %q", want.Path())
	require.Equal(t, want.IsBoolFlag, got.IsBoolFlag, "mismatched Value.IsBoolFlag() result for path %q", want.Path())
}

func RequireDataOfFlagValuesEqual(t *testing.T, want, got []*FlagValueData) {
	t.Helper()
	wantPaths, gotPaths := make([]string, len(want)), make([]string, len(got))
	for i := range want {
		wantPaths[i] = want[i].Path()
	}
	for i := range got {
		gotPaths[i] = got[i].Path()
	}
	require.Equal(t, wantPaths, gotPaths)
	for i := range want {
		RequireDataOfFlagValueEqual(t, want[i], got[i])
	}
}

func RequireDiffEqual(t *testing.T, want, got any, msgAndArgs ...any) {
	t.Helper()
	diff := cmp.Diff(want, got, protocmp.Transform())
	require.Empty(t, diff, msgAndArgs...)
}

func clone(x any) any {
	if x == nil {
		return x //nolint:gocritic // keep underlying type (if present)
	}
	return cloneValue(reflect.ValueOf(x)).Interface()
}

func cloneValue(x reflect.Value) reflect.Value {
	if m, ok := x.Interface().(proto.Message); ok {
		return reflect.ValueOf(proto.Clone(m))
	}

	switch x.Kind() { //nolint:exhaustive // Cases for types requiring special handling and used in tests only.
	case reflect.Slice:
		if x.IsNil() {
			return reflect.New(x.Type()).Elem()
		}
		c := reflect.New(x.Type())
		c.Elem().Set(reflect.MakeSlice(x.Type(), x.Len(), x.Cap()))
		c = c.Elem()
		for i, el := range x.Seq2() {
			c.Index(int(i.Int())).Set(cloneValue(el))
		}
		return c

	case reflect.Map:
		if x.IsNil() {
			return reflect.New(x.Type()).Elem()
		}
		c := reflect.New(x.Type())
		c.Elem().Set(reflect.MakeMapWithSize(x.Type(), x.Len()))
		c = c.Elem()
		for k, el := range x.Seq2() {
			c.SetMapIndex(cloneValue(k), cloneValue(el))
		}
		return c

	default:
		if !x.CanAddr() {
			c := reflect.New(x.Type())
			c.Elem().Set(x)
			return c.Elem()
		}
		return x
	}
}

func FilledTestBase() *protoflagtestv1.TestBase {
	return protoflagtestv1.TestBase_builder{
		ValueBool:        true,
		ValueInt32:       1,
		ValueSint32:      1,
		ValueSfixed32:    1,
		ValueUint32:      1,
		ValueFixed32:     1,
		ValueInt64:       1,
		ValueSint64:      1,
		ValueSfixed64:    1,
		ValueUint64:      1,
		ValueFixed64:     1,
		ValueFloat:       1,
		ValueDouble:      1,
		ValueString:      "aaa",
		ValueBytes:       []byte("aaa"),
		ValueEnum:        protoflagtestv1.TestEnum_TEST_ENUM_FIRST,
		ValueMessage:     protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build(),
		ValueMap:         map[string]string{"k": "aaa"},
		RepeatedBool:     []bool{true},
		RepeatedInt32:    []int32{1},
		RepeatedSint32:   []int32{1},
		RepeatedSfixed32: []int32{1},
		RepeatedUint32:   []uint32{1},
		RepeatedFixed32:  []uint32{1},
		RepeatedInt64:    []int64{1},
		RepeatedSint64:   []int64{1},
		RepeatedSfixed64: []int64{1},
		RepeatedUint64:   []uint64{1},
		RepeatedFixed64:  []uint64{1},
		RepeatedFloat:    []float32{1},
		RepeatedDouble:   []float64{1},
		RepeatedString:   []string{"aaa"},
		RepeatedBytes:    [][]byte{[]byte("aaa")},
		RepeatedEnum:     []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_FIRST},
		RepeatedMessage:  []*protoflagtestv1.TestMessage{protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()},
		ValueSkip:        1,
		RepeatedSkip:     []int32{1},
	}.Build()
}

func DumpJson(t *testing.T, m proto.Message) string {
	t.Helper()
	b, err := protojson.Marshal(m)
	require.NoError(t, err)
	return string(b)
}

func TestRecursiveValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		given   proto.Message
		filters []protoflag.FilterFunc
		want    []*FlagValueData
	}{
		{
			name:  "nil",
			given: nil,
			want:  []*FlagValueData(nil),
		},
		{
			name:  "message-nil",
			given: (*protoflagtestv1.TestBase)(nil),
			want:  []*FlagValueData(nil),
		},
		{
			name: "message",
			filters: []protoflag.FilterFunc{
				func(val *protoflag.Value) protoflag.FilterResult { // skip and do not descend into fields ending with "_skip"
					if p := val.Path(); len(p) > 0 && strings.HasSuffix(string(p[len(p)-1].Name()), "_skip") {
						return protoflag.SkipNoDescend
					}
					return protoflag.IncludeAndDescend
				},
			},
			given: &protoflagtestv1.TestBase{},
			want: []*FlagValueData{
				{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: "", HasMessageDescriptor: true, HasMessageType: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_bool"}, Kind: protoreflect.BoolKind, Type: "bool", Get: false, String: "", IsBoolFlag: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_int32"}, Kind: protoreflect.Int32Kind, Type: "int32", Get: int32(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sint32"}, Kind: protoreflect.Sint32Kind, Type: "int32", Get: int32(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sfixed32"}, Kind: protoreflect.Sfixed32Kind, Type: "int32", Get: int32(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_uint32"}, Kind: protoreflect.Uint32Kind, Type: "uint32", Get: uint32(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_fixed32"}, Kind: protoreflect.Fixed32Kind, Type: "uint32", Get: uint32(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_int64"}, Kind: protoreflect.Int64Kind, Type: "int64", Get: int64(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sint64"}, Kind: protoreflect.Sint64Kind, Type: "int64", Get: int64(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sfixed64"}, Kind: protoreflect.Sfixed64Kind, Type: "int64", Get: int64(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_uint64"}, Kind: protoreflect.Uint64Kind, Type: "uint64", Get: uint64(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_fixed64"}, Kind: protoreflect.Fixed64Kind, Type: "uint64", Get: uint64(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_float"}, Kind: protoreflect.FloatKind, Type: "float32", Get: float32(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_double"}, Kind: protoreflect.DoubleKind, Type: "float64", Get: float64(0), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_string"}, Kind: protoreflect.StringKind, Type: "string", Get: "", String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_bytes"}, Kind: protoreflect.BytesKind, Type: "base64", Get: []byte(nil), String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_enum"}, Kind: protoreflect.EnumKind, Type: "string", Get: protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED.Number(), String: "", HasEnumDescriptor: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_message"}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: (*protoflagtestv1.TestMessage)(nil), String: "", HasMessageDescriptor: true, HasMessageType: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_message", "protoflagtest.v1.TestMessage.value"}, Kind: protoreflect.StringKind, Type: "string", Get: "", String: ""},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_map"}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: map[string]string(nil), String: "", HasMessageDescriptor: true, IsMap: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_bool"}, Kind: protoreflect.BoolKind, Type: "bool (list)", Get: []bool(nil), String: "", IsBoolFlag: true, IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_int32"}, Kind: protoreflect.Int32Kind, Type: "int32 (list)", Get: []int32(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sint32"}, Kind: protoreflect.Sint32Kind, Type: "int32 (list)", Get: []int32(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sfixed32"}, Kind: protoreflect.Sfixed32Kind, Type: "int32 (list)", Get: []int32(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_uint32"}, Kind: protoreflect.Uint32Kind, Type: "uint32 (list)", Get: []uint32(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_fixed32"}, Kind: protoreflect.Fixed32Kind, Type: "uint32 (list)", Get: []uint32(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_int64"}, Kind: protoreflect.Int64Kind, Type: "int64 (list)", Get: []int64(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sint64"}, Kind: protoreflect.Sint64Kind, Type: "int64 (list)", Get: []int64(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sfixed64"}, Kind: protoreflect.Sfixed64Kind, Type: "int64 (list)", Get: []int64(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_uint64"}, Kind: protoreflect.Uint64Kind, Type: "uint64 (list)", Get: []uint64(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_fixed64"}, Kind: protoreflect.Fixed64Kind, Type: "uint64 (list)", Get: []uint64(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_float"}, Kind: protoreflect.FloatKind, Type: "float32 (list)", Get: []float32(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_double"}, Kind: protoreflect.DoubleKind, Type: "float64 (list)", Get: []float64(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_string"}, Kind: protoreflect.StringKind, Type: "string (list)", Get: []string(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_bytes"}, Kind: protoreflect.BytesKind, Type: "base64 (list)", Get: [][]byte(nil), String: "", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_enum"}, Kind: protoreflect.EnumKind, Type: "string (list)", Get: []protoreflect.EnumNumber(nil), String: "", HasEnumDescriptor: true, IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_message"}, Kind: protoreflect.MessageKind, Type: "JSON object (list)", Get: []proto.Message(nil), String: "", HasMessageDescriptor: true, HasMessageType: true, IsList: true},
			},
		},
		{
			name: "message-filled",
			filters: []protoflag.FilterFunc{
				func(val *protoflag.Value) protoflag.FilterResult { // skip and do not descend into fields ending with "_skip"
					if p := val.Path(); len(p) > 0 && strings.HasSuffix(string(p[len(p)-1].Name()), "_skip") {
						return protoflag.SkipNoDescend
					}
					return protoflag.IncludeAndDescend
				},
			},
			given: FilledTestBase(),
			want: []*FlagValueData{
				{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: FilledTestBase(), String: DumpJson(t, FilledTestBase()), HasMessageDescriptor: true, HasMessageType: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_bool"}, Kind: protoreflect.BoolKind, Type: "bool", Get: true, String: "true", IsBoolFlag: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_int32"}, Kind: protoreflect.Int32Kind, Type: "int32", Get: int32(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sint32"}, Kind: protoreflect.Sint32Kind, Type: "int32", Get: int32(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sfixed32"}, Kind: protoreflect.Sfixed32Kind, Type: "int32", Get: int32(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_uint32"}, Kind: protoreflect.Uint32Kind, Type: "uint32", Get: uint32(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_fixed32"}, Kind: protoreflect.Fixed32Kind, Type: "uint32", Get: uint32(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_int64"}, Kind: protoreflect.Int64Kind, Type: "int64", Get: int64(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sint64"}, Kind: protoreflect.Sint64Kind, Type: "int64", Get: int64(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_sfixed64"}, Kind: protoreflect.Sfixed64Kind, Type: "int64", Get: int64(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_uint64"}, Kind: protoreflect.Uint64Kind, Type: "uint64", Get: uint64(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_fixed64"}, Kind: protoreflect.Fixed64Kind, Type: "uint64", Get: uint64(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_float"}, Kind: protoreflect.FloatKind, Type: "float32", Get: float32(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_double"}, Kind: protoreflect.DoubleKind, Type: "float64", Get: float64(1), String: "1"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_string"}, Kind: protoreflect.StringKind, Type: "string", Get: "aaa", String: "aaa"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_bytes"}, Kind: protoreflect.BytesKind, Type: "base64", Get: []byte("aaa"), String: base64.StdEncoding.EncodeToString([]byte("aaa"))},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_enum"}, Kind: protoreflect.EnumKind, Type: "string", Get: protoflagtestv1.TestEnum_TEST_ENUM_FIRST.Number(), String: protoflagtestv1.TestEnum_TEST_ENUM_FIRST.String(), HasEnumDescriptor: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_message"}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build(), String: `{"value":"aaa"}`, HasMessageDescriptor: true, HasMessageType: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_message", "protoflagtest.v1.TestMessage.value"}, Kind: protoreflect.StringKind, Type: "string", Get: "aaa", String: "aaa"},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.value_map"}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: map[string]string{"k": "aaa"}, String: `{"k":"aaa"}`, HasMessageDescriptor: true, IsMap: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_bool"}, Kind: protoreflect.BoolKind, Type: "bool (list)", Get: []bool{true}, String: "[true]", IsBoolFlag: true, IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_int32"}, Kind: protoreflect.Int32Kind, Type: "int32 (list)", Get: []int32{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sint32"}, Kind: protoreflect.Sint32Kind, Type: "int32 (list)", Get: []int32{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sfixed32"}, Kind: protoreflect.Sfixed32Kind, Type: "int32 (list)", Get: []int32{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_uint32"}, Kind: protoreflect.Uint32Kind, Type: "uint32 (list)", Get: []uint32{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_fixed32"}, Kind: protoreflect.Fixed32Kind, Type: "uint32 (list)", Get: []uint32{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_int64"}, Kind: protoreflect.Int64Kind, Type: "int64 (list)", Get: []int64{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sint64"}, Kind: protoreflect.Sint64Kind, Type: "int64 (list)", Get: []int64{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_sfixed64"}, Kind: protoreflect.Sfixed64Kind, Type: "int64 (list)", Get: []int64{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_uint64"}, Kind: protoreflect.Uint64Kind, Type: "uint64 (list)", Get: []uint64{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_fixed64"}, Kind: protoreflect.Fixed64Kind, Type: "uint64 (list)", Get: []uint64{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_float"}, Kind: protoreflect.FloatKind, Type: "float32 (list)", Get: []float32{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_double"}, Kind: protoreflect.DoubleKind, Type: "float64 (list)", Get: []float64{1}, String: "[1]", IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_string"}, Kind: protoreflect.StringKind, Type: "string (list)", Get: []string{"aaa"}, String: `["aaa"]`, IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_bytes"}, Kind: protoreflect.BytesKind, Type: "base64 (list)", Get: [][]byte{[]byte("aaa")}, String: `["` + base64.StdEncoding.EncodeToString([]byte("aaa")) + `"]`, IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_enum"}, Kind: protoreflect.EnumKind, Type: "string (list)", Get: []protoreflect.EnumNumber{protoflagtestv1.TestEnum_TEST_ENUM_FIRST.Number()}, String: `["` + protoflagtestv1.TestEnum_TEST_ENUM_FIRST.String() + `"]`, HasEnumDescriptor: true, IsList: true},
				{PathFullNames: []string{"protoflagtest.v1.TestBase.repeated_message"}, Kind: protoreflect.MessageKind, Type: "JSON object (list)", Get: []proto.Message{protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}, String: `[{"value":"aaa"}]`, HasMessageDescriptor: true, HasMessageType: true, IsList: true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			given := proto.Clone(test.given)
			got := protoflag.Recursive(given, test.filters...)
			RequireDataOfFlagValuesEqual(t, test.want, DataOfFlagValues(got))
			RequireDiffEqual(t, test.given, given, "base message modified after call to protoflag.Recursive (-want, +got)")
		})
	}
}

func TestNewValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		given       proto.Message
		noFlagValue bool
		encoder     func(protoreflect.Value) ([]byte, error)
		decoder     func([]byte, protoreflect.Message, protoreflect.FieldDescriptor) error
		pre         *FlagValueData
		setTo       string
		setError    string
		want        *FlagValueData
	}{
		{
			name:        "nil",
			given:       nil,
			noFlagValue: true,
		},
		{
			name:        "message-nil",
			given:       (*protoflagtestv1.TestBase)(nil),
			noFlagValue: true,
		},
		{
			name:  "message",
			given: &protoflagtestv1.TestBase{},
			pre:   &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
			setTo: `{"valueString":"aaa"}`,
			want:  &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestBase_builder{ValueString: "aaa"}.Build(), String: `{"valueString":"aaa"}`, HasMessageDescriptor: true, HasMessageType: true},
		},
		{
			name:  "message-filled",
			given: protoflagtestv1.TestBase_builder{ValueBool: true}.Build(),
			pre:   &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestBase_builder{ValueBool: true}.Build(), String: `{"valueBool":true}`, HasMessageDescriptor: true, HasMessageType: true},
			setTo: `{"valueString":"aaa"}`,
			want:  &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestBase_builder{ValueString: "aaa"}.Build(), String: `{"valueString":"aaa"}`, HasMessageDescriptor: true, HasMessageType: true},
		},
		{
			name:     "message-invalid-value",
			given:    &protoflagtestv1.TestBase{},
			pre:      &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
			setTo:    `invalid`,
			setError: "invalid value",
			want:     &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
		},
		{
			name:  "message-custom-encoder",
			given: &protoflagtestv1.TestBase{},
			encoder: func(v protoreflect.Value) ([]byte, error) {
				return strconv.AppendBool(nil, v.Message().Interface().(*protoflagtestv1.TestBase).GetValueString() != ""), nil
			},
			pre:   &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: `false`, HasMessageDescriptor: true, HasMessageType: true},
			setTo: `{"valueString":"aaa"}`,
			want:  &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestBase_builder{ValueString: "aaa"}.Build(), String: `true`, HasMessageDescriptor: true, HasMessageType: true},
		},
		{
			name:    "message-custom-encoder-error",
			given:   &protoflagtestv1.TestBase{},
			encoder: func(_ protoreflect.Value) ([]byte, error) { return nil, errors.New("encoding error") },
			pre:     &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
			setTo:   `{"valueString":"aaa"}`,
			want:    &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestBase_builder{ValueString: "aaa"}.Build(), String: ``, HasMessageDescriptor: true, HasMessageType: true},
		},
		{
			name:  "message-custom-decoder",
			given: &protoflagtestv1.TestBase{},
			decoder: func(_ []byte, m protoreflect.Message, _ protoreflect.FieldDescriptor) error {
				m.Interface().(*protoflagtestv1.TestBase).SetValueString("bbb")
				return nil
			},
			pre:   &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
			setTo: `{"valueString":"aaa"}`,
			want:  &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: protoflagtestv1.TestBase_builder{ValueString: "bbb"}.Build(), String: `{"valueString":"bbb"}`, HasMessageDescriptor: true, HasMessageType: true},
		},
		{
			name:  "message-custom-decoder",
			given: &protoflagtestv1.TestBase{},
			decoder: func(_ []byte, _ protoreflect.Message, _ protoreflect.FieldDescriptor) error {
				return errors.New("decoding error")
			},
			pre:      &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
			setTo:    `{"valueString":"aaa"}`,
			setError: "decoding error",
			want:     &FlagValueData{PathFullNames: []string{}, Kind: protoreflect.MessageKind, Type: "JSON object", Get: &protoflagtestv1.TestBase{}, String: ``, HasMessageDescriptor: true, HasMessageType: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			given := proto.Clone(test.given)
			val := protoflag.New(given)
			if test.noFlagValue {
				require.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			if test.encoder != nil {
				val.SetEncoder(test.encoder)
			}
			if test.decoder != nil {
				val.SetDecoder(test.decoder)
			}

			RequireDataOfFlagValueEqual(t, test.pre, DataOfFlagValue(val))
			RequireDiffEqual(t, test.given, given, "base message modified after call to protoflag.New (-want, +got)")

			err := val.Set(test.setTo)
			if test.setError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.setError)
			} else {
				require.NoError(t, err)
			}

			RequireDataOfFlagValueEqual(t, test.want, DataOfFlagValue(val))
			RequireDiffEqual(t, test.want.Get, given, "bad base message value after call to protoflag.Value.Set (-want, +got)")
		})
	}
}

func TestZeroValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value *protoflag.Value
	}{
		{name: "nil-pointer", value: nil},
		{name: "zero-value", value: &protoflag.Value{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.value.SetEncoder(nil)
			test.value.SetDecoder(nil)
			require.Zero(t, test.value.Path())
			require.Zero(t, test.value.Kind())
			require.Zero(t, test.value.MessageDescriptor())
			require.Zero(t, test.value.MessageType())
			require.Zero(t, test.value.EnumDescriptor())
			require.Zero(t, test.value.Type()) //nolint:testifylint // use require.Zero for code self-similarity
			require.Zero(t, test.value.Get())
			require.Zero(t, test.value.GetValue())
			require.Zero(t, test.value.String()) //nolint:testifylint // use require.Zero for code self-similarity
			require.Zero(t, test.value.Set(""))  //nolint:testifylint // use require.Zero for code self-similarity
			require.Zero(t, test.value.IsBoolFlag())
		})
	}
}

func TestFlagUse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		given proto.Message
		args  []string
		want  proto.Message
	}{
		{
			name:  "no-args",
			given: &protoflagtestv1.TestBase{},
			args:  []string{},
			want:  &protoflagtestv1.TestBase{},
		},
		{
			name:  "bool/false",
			given: protoflagtestv1.TestBase_builder{ValueBool: true}.Build(),
			args:  []string{"--valueBool=false"},
			want:  protoflagtestv1.TestBase_builder{ValueBool: false}.Build(),
		},
		{
			name:  "bool/true",
			given: protoflagtestv1.TestBase_builder{ValueBool: false}.Build(),
			args:  []string{"--valueBool=true"},
			want:  protoflagtestv1.TestBase_builder{ValueBool: true}.Build(),
		},
		{
			name:  "bool/true-implicit",
			given: protoflagtestv1.TestBase_builder{ValueBool: false}.Build(),
			args:  []string{"--valueBool"},
			want:  protoflagtestv1.TestBase_builder{ValueBool: true}.Build(),
		},
		{
			name:  "int32/0",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=0"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "sint32/0",
			given: protoflagtestv1.TestBase_builder{ValueSint32: 2}.Build(),
			args:  []string{"--valueSint32=0"},
			want:  protoflagtestv1.TestBase_builder{ValueSint32: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "sfixed32/0",
			given: protoflagtestv1.TestBase_builder{ValueSfixed32: 2}.Build(),
			args:  []string{"--valueSfixed32=0"},
			want:  protoflagtestv1.TestBase_builder{ValueSfixed32: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "uint32/0",
			given: protoflagtestv1.TestBase_builder{ValueUint32: 2}.Build(),
			args:  []string{"--valueUint32=0"},
			want:  protoflagtestv1.TestBase_builder{ValueUint32: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "fixed32/0",
			given: protoflagtestv1.TestBase_builder{ValueFixed32: 2}.Build(),
			args:  []string{"--valueFixed32=0"},
			want:  protoflagtestv1.TestBase_builder{ValueFixed32: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "int64/0",
			given: protoflagtestv1.TestBase_builder{ValueInt64: 2}.Build(),
			args:  []string{"--valueInt64=0"},
			want:  protoflagtestv1.TestBase_builder{ValueInt64: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "sint64/0",
			given: protoflagtestv1.TestBase_builder{ValueSint64: 2}.Build(),
			args:  []string{"--valueSint64=0"},
			want:  protoflagtestv1.TestBase_builder{ValueSint64: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "sfixed64/0",
			given: protoflagtestv1.TestBase_builder{ValueSfixed64: 2}.Build(),
			args:  []string{"--valueSfixed64=0"},
			want:  protoflagtestv1.TestBase_builder{ValueSfixed64: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "uint64/0",
			given: protoflagtestv1.TestBase_builder{ValueUint64: 2}.Build(),
			args:  []string{"--valueUint64=0"},
			want:  protoflagtestv1.TestBase_builder{ValueUint64: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "fixed64/0",
			given: protoflagtestv1.TestBase_builder{ValueFixed64: 2}.Build(),
			args:  []string{"--valueFixed64=0"},
			want:  protoflagtestv1.TestBase_builder{ValueFixed64: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "float/0",
			given: protoflagtestv1.TestBase_builder{ValueFloat: 2}.Build(),
			args:  []string{"--valueFloat=0"},
			want:  protoflagtestv1.TestBase_builder{ValueFloat: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "double/0",
			given: protoflagtestv1.TestBase_builder{ValueDouble: 2}.Build(),
			args:  []string{"--valueDouble=0"},
			want:  protoflagtestv1.TestBase_builder{ValueDouble: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "double/0",
			given: protoflagtestv1.TestBase_builder{ValueDouble: 2}.Build(),
			args:  []string{"--valueDouble=0"},
			want:  protoflagtestv1.TestBase_builder{ValueDouble: 0}.Build(),
		},
		{
			name:  "int32/1",
			given: protoflagtestv1.TestBase_builder{ValueInt32: 2}.Build(),
			args:  []string{"--valueInt32=1"},
			want:  protoflagtestv1.TestBase_builder{ValueInt32: 1}.Build(),
		},
		{
			name:  "string/empty",
			given: protoflagtestv1.TestBase_builder{ValueString: "bbb"}.Build(),
			args:  []string{"--valueString="},
			want:  protoflagtestv1.TestBase_builder{ValueString: ""}.Build(),
		},
		{
			name:  "string/non-empty",
			given: protoflagtestv1.TestBase_builder{ValueString: "bbb"}.Build(),
			args:  []string{"--valueString=aaa"},
			want:  protoflagtestv1.TestBase_builder{ValueString: "aaa"}.Build(),
		},
		{
			name:  "bytes/empty",
			given: protoflagtestv1.TestBase_builder{ValueBytes: []byte("bbb")}.Build(),
			args:  []string{"--valueBytes="},
			want:  protoflagtestv1.TestBase_builder{ValueBytes: []byte{}}.Build(),
		},
		{
			name:  "bytes/non-empty",
			given: protoflagtestv1.TestBase_builder{ValueBytes: []byte("bbb")}.Build(),
			args:  []string{"--valueBytes=" + base64.StdEncoding.EncodeToString([]byte("aaa"))},
			want:  protoflagtestv1.TestBase_builder{ValueBytes: []byte("aaa")}.Build(),
		},
		{
			name:  "enum/unspecified",
			given: protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_SECOND}.Build(),
			args:  []string{"--valueEnum=" + protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED.String()},
			want:  protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED}.Build(),
		},
		{
			name:  "enum/0",
			given: protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_SECOND}.Build(),
			args:  []string{"--valueEnum=" + strconv.Itoa(int(protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED.Number()))},
			want:  protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED}.Build(),
		},
		{
			name:  "enum/specified",
			given: protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_SECOND}.Build(),
			args:  []string{"--valueEnum=" + protoflagtestv1.TestEnum_TEST_ENUM_FIRST.String()},
			want:  protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_FIRST}.Build(),
		},
		{
			name:  "enum/1",
			given: protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_SECOND}.Build(),
			args:  []string{"--valueEnum=" + strconv.Itoa(int(protoflagtestv1.TestEnum_TEST_ENUM_FIRST.Number()))},
			want:  protoflagtestv1.TestBase_builder{ValueEnum: protoflagtestv1.TestEnum_TEST_ENUM_FIRST}.Build(),
		},
		{
			name:  "message/non-empty-object-init",
			given: protoflagtestv1.TestBase_builder{}.Build(),
			args:  []string{`--valueMessage={"value":"aaa"}`},
			want:  protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}.Build(),
		},
		{
			name:  "message/empty-object",
			given: protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build()}.Build(),
			args:  []string{"--valueMessage={}"},
			want:  protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{}.Build()}.Build(),
		},
		{
			name:  "message/non-empty-object",
			given: protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build()}.Build(),
			args:  []string{`--valueMessage={"value":"aaa"}`},
			want:  protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}.Build(),
		},
		{
			name:  "message/field/non-empty-init",
			given: protoflagtestv1.TestBase_builder{}.Build(),
			args:  []string{`--valueMessage.value=aaa`},
			want:  protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}.Build(),
		},
		{
			name:  "message/field/empty",
			given: protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build()}.Build(),
			args:  []string{"--valueMessage.value="},
			want:  protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: ""}.Build()}.Build(),
		},
		{
			name:  "message/field/non-empty",
			given: protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build()}.Build(),
			args:  []string{`--valueMessage.value=aaa`},
			want:  protoflagtestv1.TestBase_builder{ValueMessage: protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}.Build(),
		},
		{
			name:  "map/non-empty-object-init",
			given: protoflagtestv1.TestBase_builder{}.Build(),
			args:  []string{`--valueMap={"k":"aaa"}`},
			want:  protoflagtestv1.TestBase_builder{ValueMap: map[string]string{"k": "aaa"}}.Build(),
		},
		{
			name:  "map/empty-object",
			given: protoflagtestv1.TestBase_builder{ValueMap: map[string]string{"k": "bbb"}}.Build(),
			args:  []string{"--valueMap={}"},
			want:  protoflagtestv1.TestBase_builder{ValueMap: map[string]string{}}.Build(),
		},
		{
			name:  "map/non-empty-object",
			given: protoflagtestv1.TestBase_builder{ValueMap: map[string]string{"k": "bbb"}}.Build(),
			args:  []string{`--valueMap={"k":"aaa"}`},
			want:  protoflagtestv1.TestBase_builder{ValueMap: map[string]string{"k": "aaa"}}.Build(),
		},
		{
			name:  "repeated-bool/true-init",
			given: protoflagtestv1.TestBase_builder{RepeatedBool: nil}.Build(),
			args:  []string{"--repeatedBool=true"},
			want:  protoflagtestv1.TestBase_builder{RepeatedBool: []bool{true}}.Build(),
		},
		{
			name:  "repeated-bool/false",
			given: protoflagtestv1.TestBase_builder{RepeatedBool: []bool{true}}.Build(),
			args:  []string{"--repeatedBool=false"},
			want:  protoflagtestv1.TestBase_builder{RepeatedBool: []bool{true, false}}.Build(),
		},
		{
			name:  "repeated-bool/true",
			given: protoflagtestv1.TestBase_builder{RepeatedBool: []bool{false}}.Build(),
			args:  []string{"--repeatedBool=true"},
			want:  protoflagtestv1.TestBase_builder{RepeatedBool: []bool{false, true}}.Build(),
		},
		{
			name:  "repeated-bool/true-implicit",
			given: protoflagtestv1.TestBase_builder{RepeatedBool: []bool{false}}.Build(),
			args:  []string{"--repeatedBool"},
			want:  protoflagtestv1.TestBase_builder{RepeatedBool: []bool{false, true}}.Build(),
		},
		{
			name:  "repeated-int32/1-init",
			given: protoflagtestv1.TestBase_builder{RepeatedInt32: nil}.Build(),
			args:  []string{"--repeatedInt32=1"},
			want:  protoflagtestv1.TestBase_builder{RepeatedInt32: []int32{1}}.Build(),
		},
		{
			name:  "repeated-int32/0",
			given: protoflagtestv1.TestBase_builder{RepeatedInt32: []int32{2}}.Build(),
			args:  []string{"--repeatedInt32=0"},
			want:  protoflagtestv1.TestBase_builder{RepeatedInt32: []int32{2, 0}}.Build(),
		},
		{
			name:  "repeated-int32/1",
			given: protoflagtestv1.TestBase_builder{RepeatedInt32: []int32{2}}.Build(),
			args:  []string{"--repeatedInt32=1"},
			want:  protoflagtestv1.TestBase_builder{RepeatedInt32: []int32{2, 1}}.Build(),
		},
		{
			name:  "repeated-string/non-empty-init",
			given: protoflagtestv1.TestBase_builder{RepeatedString: nil}.Build(),
			args:  []string{"--repeatedString=aaa"},
			want:  protoflagtestv1.TestBase_builder{RepeatedString: []string{"aaa"}}.Build(),
		},
		{
			name:  "repeated-string/empty",
			given: protoflagtestv1.TestBase_builder{RepeatedString: []string{"bbb"}}.Build(),
			args:  []string{"--repeatedString="},
			want:  protoflagtestv1.TestBase_builder{RepeatedString: []string{"bbb", ""}}.Build(),
		},
		{
			name:  "repeated-string/non-empty",
			given: protoflagtestv1.TestBase_builder{RepeatedString: []string{"bbb"}}.Build(),
			args:  []string{"--repeatedString=aaa"},
			want:  protoflagtestv1.TestBase_builder{RepeatedString: []string{"bbb", "aaa"}}.Build(),
		},
		{
			name:  "repeated-bytes/non-empty-init",
			given: protoflagtestv1.TestBase_builder{RepeatedBytes: nil}.Build(),
			args:  []string{"--repeatedBytes=" + base64.StdEncoding.EncodeToString([]byte("aaa"))},
			want:  protoflagtestv1.TestBase_builder{RepeatedBytes: [][]byte{[]byte("aaa")}}.Build(),
		},
		{
			name:  "repeated-bytes/empty",
			given: protoflagtestv1.TestBase_builder{RepeatedBytes: [][]byte{[]byte("bbb")}}.Build(),
			args:  []string{"--repeatedBytes="},
			want:  protoflagtestv1.TestBase_builder{RepeatedBytes: [][]byte{[]byte("bbb"), {}}}.Build(),
		},
		{
			name:  "repeated-bytes/non-empty",
			given: protoflagtestv1.TestBase_builder{RepeatedBytes: [][]byte{[]byte("bbb")}}.Build(),
			args:  []string{"--repeatedBytes=" + base64.StdEncoding.EncodeToString([]byte("aaa"))},
			want:  protoflagtestv1.TestBase_builder{RepeatedBytes: [][]byte{[]byte("bbb"), []byte("aaa")}}.Build(),
		},
		{
			name:  "repeated-enum/specified-init",
			given: protoflagtestv1.TestBase_builder{RepeatedEnum: nil}.Build(),
			args:  []string{"--repeatedEnum=" + protoflagtestv1.TestEnum_TEST_ENUM_FIRST.String()},
			want:  protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_FIRST}}.Build(),
		},
		{
			name:  "repeated-enum/unspecified",
			given: protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND}}.Build(),
			args:  []string{"--repeatedEnum=" + protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED.String()},
			want:  protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND, protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED}}.Build(),
		},
		{
			name:  "repeated-enum/0",
			given: protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND}}.Build(),
			args:  []string{"--repeatedEnum=" + strconv.Itoa(int(protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED.Number()))},
			want:  protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND, protoflagtestv1.TestEnum_TEST_ENUM_UNSPECIFIED}}.Build(),
		},
		{
			name:  "repeated-enum/specified",
			given: protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND}}.Build(),
			args:  []string{"--repeatedEnum=" + protoflagtestv1.TestEnum_TEST_ENUM_FIRST.String()},
			want:  protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND, protoflagtestv1.TestEnum_TEST_ENUM_FIRST}}.Build(),
		},
		{
			name:  "repeated-enum/1",
			given: protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND}}.Build(),
			args:  []string{"--repeatedEnum=" + strconv.Itoa(int(protoflagtestv1.TestEnum_TEST_ENUM_FIRST.Number()))},
			want:  protoflagtestv1.TestBase_builder{RepeatedEnum: []protoflagtestv1.TestEnum{protoflagtestv1.TestEnum_TEST_ENUM_SECOND, protoflagtestv1.TestEnum_TEST_ENUM_FIRST}}.Build(),
		},
		{
			name:  "repeated-message/non-empty-object-init",
			given: protoflagtestv1.TestBase_builder{RepeatedMessage: nil}.Build(),
			args:  []string{`--repeatedMessage={"value":"aaa"}`},
			want:  protoflagtestv1.TestBase_builder{RepeatedMessage: []*protoflagtestv1.TestMessage{protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}}.Build(),
		},
		{
			name:  "repeated-message/empty-object",
			given: protoflagtestv1.TestBase_builder{RepeatedMessage: []*protoflagtestv1.TestMessage{protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build()}}.Build(),
			args:  []string{"--repeatedMessage={}"},
			want:  protoflagtestv1.TestBase_builder{RepeatedMessage: []*protoflagtestv1.TestMessage{protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build(), protoflagtestv1.TestMessage_builder{}.Build()}}.Build(),
		},
		{
			name:  "repeated-message/non-empty-object",
			given: protoflagtestv1.TestBase_builder{RepeatedMessage: []*protoflagtestv1.TestMessage{protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build()}}.Build(),
			args:  []string{`--repeatedMessage={"value":"aaa"}`},
			want:  protoflagtestv1.TestBase_builder{RepeatedMessage: []*protoflagtestv1.TestMessage{protoflagtestv1.TestMessage_builder{Value: "bbb"}.Build(), protoflagtestv1.TestMessage_builder{Value: "aaa"}.Build()}}.Build(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			given := proto.Clone(test.given)
			fs := flag.NewFlagSet("", flag.ContinueOnError)
			for _, val := range protoflag.Recursive(given) {
				fs.Var(val, protoflag.JSONName(val.Path()), fmt.Sprintf("usage of --%s flag", protoflag.JSONName(val.Path())))
			}
			err := fs.Parse(test.args)
			require.NoError(t, err, "flag.FlagSet.Parse failed")
			RequireDiffEqual(t, test.want, given, "bad base message value after parsing arguments (-want, +got)")
		})
	}
}
