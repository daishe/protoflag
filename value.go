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

package protoflag

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/daishe/jsonflag"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// New returns new flag value for the provided proto message. It returns nil if the message cannot be used as flag value.
func New(base proto.Message) *Value {
	if base == nil {
		return nil
	}
	v := base.ProtoReflect()
	if !v.IsValid() {
		return nil
	}
	return newValue(v, nil)
}

// FilterFunc is a function that can be used to decide if a flag value should be included in recursive results as well as if flag values finding should descend into the sub-values.
type FilterFunc func(*Value) FilterResult

// FilterResult is a bit field holding filtering decision.
type FilterResult = jsonflag.FilterResult

const (
	IncludeAndDescend = jsonflag.IncludeAndDescend // indicates that the given flag value should be included, and that recursive value retrieval should descend into sub-values
	IncludeNoDescend  = jsonflag.IncludeNoDescend  // indicates that the given flag value should be included, but that recursive value retrieval should NOT descend into sub-values
	SkipAndDescend    = jsonflag.SkipAndDescend    // indicates that the given flag value should be skipped, but that recursive value retrieval should descend into sub-values
	SkipNoDescend     = jsonflag.SkipNoDescend     // indicates that the given flag value should be skipped, and that recursive value retrieval should NOT descend into sub-values
)

// ShouldInclude informs if the provided filter result indicates that the given flag value should be included.
func ShouldInclude(r FilterResult) bool {
	return jsonflag.ShouldInclude(r)
}

// ShouldDescend informs if the provided filter result indicates that the further recursive values finding should descend into sub-values.
func ShouldDescend(r FilterResult) bool {
	return jsonflag.ShouldDescend(r)
}

// Filter applies all filter function to the provided flag value and returns filtering decision.
func Filter(val *Value, filters ...FilterFunc) (r FilterResult) {
	if val == nil {
		return SkipNoDescend
	}
	for _, cond := range filters {
		r |= cond(val)
	}
	return r
}

// Recursive returns set of flag values for the provided proto message and all fields within, recursively, according to the provided filters. Function silently skips all the fields that cannot be used as flag values.
func Recursive(base proto.Message, filters ...FilterFunc) []*Value {
	if base == nil {
		return nil
	}
	v := base.ProtoReflect()
	if !v.IsValid() {
		return nil
	}
	return recursive(v, nil, filters)
}

func recursive(base protoreflect.Message, fields []protoreflect.FieldDescriptor, filters []FilterFunc) []*Value {
	values := []*Value(nil)
	val := newValue(base, slices.Clone(fields))
	filterResult := Filter(val, filters...)
	if ShouldInclude(filterResult) {
		values = append(values, val)
	}
	if !ShouldDescend(filterResult) {
		return values
	}

	md := base.Descriptor()
	if len(fields) > 0 {
		fd := fields[len(fields)-1]
		if fd.Kind() != protoreflect.MessageKind || fd.IsList() || fd.IsMap() {
			return values
		}
		md = fd.Message()
	}

	mfds := md.Fields()
	for i := range mfds.Len() {
		field := mfds.Get(i)
		values = append(values, recursive(base, append(fields, field), filters)...)
	}
	return values
}

func newValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	if len(fields) > 0 && fields[len(fields)-1].IsList() {
		return newRepeatedValue(base, fields)
	}
	if len(fields) > 0 && fields[len(fields)-1].IsMap() {
		return newMapValue(base, fields)
	}

	k := protoreflect.MessageKind
	if len(fields) > 0 {
		k = fields[len(fields)-1].Kind()
	}
	//nolint:dupl // This is not duplicated!
	switch k { //nolint:exhaustive // Cases for supported kinds only.
	case protoreflect.BoolKind:
		return newBoolValue(base, fields)
	case protoreflect.EnumKind:
		return newEnumValue(base, fields)
	case protoreflect.Int32Kind:
		return newInt32Value(base, fields)
	case protoreflect.Sint32Kind:
		return newInt32Value(base, fields)
	case protoreflect.Uint32Kind:
		return newUint32Value(base, fields)
	case protoreflect.Int64Kind:
		return newInt64Value(base, fields)
	case protoreflect.Sint64Kind:
		return newInt64Value(base, fields)
	case protoreflect.Uint64Kind:
		return newUint64Value(base, fields)
	case protoreflect.Sfixed32Kind:
		return newInt32Value(base, fields)
	case protoreflect.Fixed32Kind:
		return newUint32Value(base, fields)
	case protoreflect.FloatKind:
		return newFloat32Value(base, fields)
	case protoreflect.Sfixed64Kind:
		return newInt64Value(base, fields)
	case protoreflect.Fixed64Kind:
		return newUint64Value(base, fields)
	case protoreflect.DoubleKind:
		return newFloat64Value(base, fields)
	case protoreflect.StringKind:
		return newStringValue(base, fields)
	case protoreflect.BytesKind:
		return newBytesValue(base, fields)
	case protoreflect.MessageKind:
		return newMessageValue(base, fields)
	}
	return nil
}

func newRepeatedValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	k := fields[len(fields)-1].Kind()
	//nolint:dupl // This is not duplicated!
	switch k { //nolint:exhaustive // Cases for supported kinds only.
	case protoreflect.BoolKind:
		return newRepeatedBoolValue(base, fields)
	case protoreflect.EnumKind:
		return newRepeatedEnumValue(base, fields)
	case protoreflect.Int32Kind:
		return newRepeatedInt32Value(base, fields)
	case protoreflect.Sint32Kind:
		return newRepeatedInt32Value(base, fields)
	case protoreflect.Uint32Kind:
		return newRepeatedUint32Value(base, fields)
	case protoreflect.Int64Kind:
		return newRepeatedInt64Value(base, fields)
	case protoreflect.Sint64Kind:
		return newRepeatedInt64Value(base, fields)
	case protoreflect.Uint64Kind:
		return newRepeatedUint64Value(base, fields)
	case protoreflect.Sfixed32Kind:
		return newRepeatedInt32Value(base, fields)
	case protoreflect.Fixed32Kind:
		return newRepeatedUint32Value(base, fields)
	case protoreflect.FloatKind:
		return newRepeatedFloat32Value(base, fields)
	case protoreflect.Sfixed64Kind:
		return newRepeatedInt64Value(base, fields)
	case protoreflect.Fixed64Kind:
		return newRepeatedUint64Value(base, fields)
	case protoreflect.DoubleKind:
		return newRepeatedFloat64Value(base, fields)
	case protoreflect.StringKind:
		return newRepeatedStringValue(base, fields)
	case protoreflect.BytesKind:
		return newRepeatedBytesValue(base, fields)
	case protoreflect.MessageKind:
		return newRepeatedMessageValue(base, fields)
	}
	return nil
}

// Value is a value of a flag for some proto message or its field.
type Value struct {
	base     protoreflect.Message
	fields   []protoreflect.FieldDescriptor
	typeName string
	stringFn func(*Value) string
	encodeFn func(protoreflect.Value) ([]byte, error)
	setFn    func(*Value, string) error
	decodeFn func([]byte, protoreflect.Message, protoreflect.FieldDescriptor) error
	isBool   bool
}

func (val *Value) fieldDescriptor() protoreflect.FieldDescriptor {
	if len(val.fields) > 0 {
		return val.fields[len(val.fields)-1]
	}
	return nil
}

// Path returns list of field descriptor from the base message up to the field associated with the provided value. It returns empty list if the value is associated with the base message itself.
func (val *Value) Path() []protoreflect.FieldDescriptor {
	if !val.isInitialized() {
		return nil
	}
	return val.fields
}

// Kind returns kind of the type described by the value.
func (val *Value) Kind() protoreflect.Kind {
	if !val.isInitialized() {
		return 0
	}
	fd := val.fieldDescriptor()
	if fd == nil {
		return protoreflect.MessageKind
	}
	return fd.Kind()
}

// MessageDescriptor returns message descriptor of the message associated with the provided value, or nil, if the value is not associated with a message.
func (val *Value) MessageDescriptor() protoreflect.MessageDescriptor {
	if !val.isInitialized() {
		return nil
	}
	fd := val.fieldDescriptor()
	if fd == nil {
		return val.base.Descriptor()
	}
	return fd.Message()
}

// MessageType returns message type of the message associated with the provided value, or nil, if the value is not associated with a message.
func (val *Value) MessageType() protoreflect.MessageType {
	if !val.isInitialized() {
		return nil
	}
	fd := val.fieldDescriptor()
	if fd == nil {
		return val.GetValue().Message().Type()
	}
	switch {
	case fd.IsList() && fd.Kind() == protoreflect.MessageKind:
		li := val.GetValue().List()
		if li.Len() > 0 {
			return li.Get(0).Message().Type()
		}
		return li.NewElement().Message().Type()
	case fd.IsMap():
		return nil
	case fd.Kind() == protoreflect.MessageKind:
		return val.GetValue().Message().Type()
	}
	return nil
}

// EnumDescriptor returns enum descriptor of the enum associated with the provided value, or nil, if the value is not associated with an enum.
func (val *Value) EnumDescriptor() protoreflect.EnumDescriptor {
	if !val.isInitialized() {
		return nil
	}
	fd := val.fieldDescriptor()
	if fd == nil {
		return nil
	}
	return fd.Enum()
}

// IsList returns true of the type associated with the provided value is a list.
func (val *Value) IsList() bool {
	if !val.isInitialized() {
		return false
	}
	fd := val.fieldDescriptor()
	if fd == nil {
		return false
	}
	return fd.IsList()
}

// IsMap returns true of the type associated with the provided value is a map.
func (val *Value) IsMap() bool {
	if !val.isInitialized() {
		return false
	}
	fd := val.fieldDescriptor()
	if fd == nil {
		return false
	}
	return fd.IsMap()
}

// Type returns type name that should be displayed when using the provided value as a flag.
func (val *Value) Type() string {
	if !val.isInitialized() {
		return ""
	}
	return val.typeName
}

// Get returns current underlying value of the provided value. The actual data returned depends on underlying type associated with the value:
//   - for scalars function returns their associated Go values directly,
//   - for enums function returns enum number (protoreflect.EnumNumber type)
//   - for messages function returns their values directly,
//   - for lists of scalars function returns slices of their associated Go values,
//   - for lists of enums function returns slices of enum numbers (protoreflect.EnumNumber type),
//   - for lists of messages function returns slices of proto.Messages,
//   - for maps of scalars function returns maps of their associated Go values,
//   - for maps of enums function returns maps of enum numbers (protoreflect.EnumNumber type),
//   - for maps of messages function returns maps of proto.Messages.
func (val *Value) Get() any {
	v := val.GetValue()
	if !v.IsValid() {
		return nil
	}
	fd := val.fieldDescriptor()
	if fd == nil { // base message
		return v.Message().Interface()
	}
	return interfaceOfField(v, fd)
}

// GetValue returns current underlying protoreflect value of the provided value.
func (val *Value) GetValue() protoreflect.Value {
	if !val.isInitialized() {
		return protoreflect.Value{}
	}
	return getValueByFields(val.base, val.fields)
}

func (val *Value) getParent() protoreflect.Message {
	if !val.isInitialized() {
		return nil
	}
	if len(val.fields) == 0 {
		return nil
	}
	return getValueByFields(val.base, val.fields[:len(val.fields)-1]).Message()
}

// String returns string representation of the underlying value.
func (val *Value) String() string {
	if !val.isInitialized() {
		return ""
	}
	if val.encodeFn != nil {
		b, err := val.encodeFn(val.GetValue())
		if err != nil {
			return ""
		}
		return string(b)
	}
	return val.stringFn(val)
}

// Set converts the provided string setting the underlying value.
func (val *Value) Set(to string) error {
	if !val.isInitialized() {
		return nil
	}
	if val.decodeFn != nil {
		if len(val.fields) == 0 {
			return val.decodeFn([]byte(to), val.base, nil)
		}
		return val.decodeFn([]byte(to), mutableValueByFields(val.base, val.fields[:len(val.fields)-1]).Message(), val.fields[len(val.fields)-1])
	}
	return val.setFn(val, to)
}

func (val *Value) set(to protoreflect.Value) {
	setValueByFields(val.base, val.fields, to)
}

func (val *Value) append(new protoreflect.Value) {
	appendValueByFields(val.base, val.fields, new)
}

// SetEncoder allows setting custom encoder of the underlying value.
func (val *Value) SetEncoder(fn func(protoreflect.Value) ([]byte, error)) {
	if !val.isInitialized() {
		return
	}
	val.encodeFn = fn
}

// SetDecoder allows setting custom decoder of the underlying value.
func (val *Value) SetDecoder(fn func([]byte, protoreflect.Message, protoreflect.FieldDescriptor) error) {
	if !val.isInitialized() {
		return
	}
	val.decodeFn = fn
}

// IsBoolFlag informs if the underlying value is bool-like (bool field or repeated bool field).
func (val *Value) IsBoolFlag() bool {
	if !val.isInitialized() {
		return false
	}
	return val.isBool
}

func (val *Value) isInitialized() bool {
	return val != nil && val.base != nil
}

func getValueByFields(v protoreflect.Message, fields []protoreflect.FieldDescriptor) protoreflect.Value {
	if len(fields) == 0 {
		return protoreflect.ValueOfMessage(v)
	}
	for _, fd := range fields[:len(fields)-1] {
		v = v.Get(fd).Message()
	}
	return v.Get(fields[len(fields)-1])
}

func mutableValueByFields(v protoreflect.Message, fields []protoreflect.FieldDescriptor) protoreflect.Value {
	if len(fields) == 0 {
		return protoreflect.ValueOfMessage(v)
	}
	for _, fd := range fields[:len(fields)-1] {
		v = v.Mutable(fd).Message()
	}
	last := fields[len(fields)-1]
	if !v.Has(last) {
		v.Set(last, v.NewField(last))
	}
	return v.Get(last)
}

func setValueByFields(v protoreflect.Message, fields []protoreflect.FieldDescriptor, to protoreflect.Value) {
	if len(fields) == 0 {
		to := to.Message()
		fields := v.Descriptor().Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			if !to.Has(fd) {
				v.Clear(fd)
				continue
			}
			v.Set(fd, to.Get(fd))
		}
		return
	}
	for _, fd := range fields[:len(fields)-1] {
		v = v.Mutable(fd).Message()
	}
	v.Set(fields[len(fields)-1], to)
}

func appendValueByFields(v protoreflect.Message, fields []protoreflect.FieldDescriptor, new protoreflect.Value) {
	if len(fields) == 0 {
		panic("protoflag: cannot append to message")
	}
	for _, fd := range fields[:len(fields)-1] {
		v = v.Mutable(fd).Message()
	}
	v.Mutable(fields[len(fields)-1]).List().Append(new)
}

func reflectTypeOfKind(k protoreflect.Kind) reflect.Type {
	switch k { //nolint:exhaustive // cases for supported kinds only
	case protoreflect.BoolKind:
		return reflect.TypeFor[bool]()
	case protoreflect.EnumKind:
		return reflect.TypeFor[protoreflect.EnumNumber]()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return reflect.TypeFor[int32]()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return reflect.TypeFor[uint32]()
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return reflect.TypeFor[int64]()
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return reflect.TypeFor[uint64]()
	case protoreflect.FloatKind:
		return reflect.TypeFor[float32]()
	case protoreflect.DoubleKind:
		return reflect.TypeFor[float64]()
	case protoreflect.StringKind:
		return reflect.TypeFor[string]()
	case protoreflect.BytesKind:
		return reflect.TypeFor[[]byte]()
	case protoreflect.MessageKind:
		return reflect.TypeFor[proto.Message]()
	}
	return nil
}

func reflectTypeOfField(fd protoreflect.FieldDescriptor) reflect.Type {
	if fd == nil {
		return nil
	}
	if fd.IsList() {
		return reflect.SliceOf(reflectTypeOfKind(fd.Kind()))
	}
	if fd.IsMap() {
		return reflect.MapOf(reflectTypeOfKind(fd.MapKey().Kind()), reflectTypeOfKind(fd.MapValue().Kind()))
	}
	return reflectTypeOfKind(fd.Kind())
}

func interfaceOfKind(v protoreflect.Value, k protoreflect.Kind) any {
	if k == protoreflect.MessageKind {
		return v.Message().Interface()
	}
	return v.Interface()
}

func interfaceOfField(v protoreflect.Value, fd protoreflect.FieldDescriptor) any {
	if !v.IsValid() {
		return reflect.Zero(reflectTypeOfField(fd)).Interface()
	}
	x := v.Interface()
	switch x := x.(type) {
	case protoreflect.Message:
		return x.Interface()

	case protoreflect.List:
		if !x.IsValid() {
			return reflect.Zero(reflectTypeOfField(fd)).Interface()
		}
		k := fd.Kind()
		s := reflect.MakeSlice(reflectTypeOfField(fd), x.Len(), x.Len())
		for i := range x.Len() {
			s.Index(i).Set(reflect.ValueOf(interfaceOfKind(x.Get(i), k)))
		}
		return s.Interface()

	case protoreflect.Map:
		if !x.IsValid() {
			return reflect.Zero(reflectTypeOfField(fd)).Interface()
		}
		kk, vk := fd.MapKey().Kind(), fd.MapValue().Kind()
		m := reflect.MakeMapWithSize(reflectTypeOfField(fd), x.Len())
		for k, v := range x.Range {
			m.SetMapIndex(
				reflect.ValueOf(interfaceOfKind(k.Value(), kk)),
				reflect.ValueOf(interfaceOfKind(v, vk)),
			)
		}
		return m.Interface()
	}
	return x
}

func newBoolValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "bool",
		stringFn: boolValueString,
		setFn:    boolValueSet,
		isBool:   true,
	}
}

func boolValueString(val *Value) string {
	v := val.GetValue().Bool()
	if !v {
		return ""
	}
	return strconv.FormatBool(v)
}

func boolValueSet(val *Value, to string) error {
	v, err := strconv.ParseBool(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfBool(v))
	return nil
}

func newEnumValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "string",
		stringFn: enumValueString,
		setFn:    enumValueSet,
	}
}

func enumValueString(val *Value) string {
	v := val.GetValue().Enum()
	if v == 0 {
		return ""
	}
	desc := val.EnumDescriptor().Values().ByNumber(v)
	if desc != nil {
		return string(desc.Name())
	}
	return strconv.FormatInt(int64(v), 10)
}

func enumValueSet(val *Value, to string) error {
	desc := val.EnumDescriptor().Values().ByName(protoreflect.Name(to))
	if desc != nil {
		val.set(protoreflect.ValueOfEnum(desc.Number()))
		return nil
	}
	v, err := strconv.ParseInt(to, 10, 32)
	if err != nil {
		return fmt.Errorf("unknown enum value %q", to)
	}
	val.set(protoreflect.ValueOfEnum(protoreflect.EnumNumber(v)))
	return nil
}

func intValueString(val *Value) string {
	v := val.GetValue().Int()
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

func newInt32Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "int32",
		stringFn: intValueString,
		setFn:    int32ValueSet,
	}
}

func int32ValueSet(val *Value, to string) error {
	v, err := parseInt32(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfInt32(v))
	return nil
}

func newInt64Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "int64",
		stringFn: intValueString,
		setFn:    int64ValueSet,
	}
}

func int64ValueSet(val *Value, to string) error {
	v, err := parseInt64(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfInt64(v))
	return nil
}

func uintValueString(val *Value) string {
	v := val.GetValue().Uint()
	if v == 0 {
		return ""
	}
	return strconv.FormatUint(v, 10)
}

func newUint32Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "uint32",
		stringFn: uintValueString,
		setFn:    uint32ValueSet,
	}
}

func uint32ValueSet(val *Value, to string) error {
	v, err := parseUint32(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfUint32(v))
	return nil
}

func newUint64Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "uint64",
		stringFn: uintValueString,
		setFn:    uint64ValueSet,
	}
}

func uint64ValueSet(val *Value, to string) error {
	v, err := parseUint64(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfUint64(v))
	return nil
}

func newFloat32Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "float32",
		stringFn: float32ValueString,
		setFn:    float32ValueSet,
	}
}

func float32ValueString(val *Value) string {
	v := val.GetValue().Float()
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'g', -1, 32)
}

func float32ValueSet(val *Value, to string) error {
	v, err := parseFloat32(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfFloat32(v))
	return nil
}

func newFloat64Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "float64",
		stringFn: float64ValueString,
		setFn:    float64ValueSet,
	}
}

func float64ValueString(val *Value) string {
	v := val.GetValue().Float()
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func float64ValueSet(val *Value, to string) error {
	v, err := parseFloat64(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfFloat64(v))
	return nil
}

func newStringValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "string",
		stringFn: stringValueString,
		setFn:    stringValueSet,
	}
}

func stringValueString(val *Value) string {
	return val.GetValue().String()
}

func stringValueSet(val *Value, to string) error {
	val.set(protoreflect.ValueOfString(to))
	return nil
}

func newBytesValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "base64",
		stringFn: bytesValueString,
		setFn:    bytesValueSet,
	}
}

func bytesValueString(val *Value) string {
	v := val.GetValue().Bytes()
	if len(v) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(v)
}

func bytesValueSet(val *Value, to string) error {
	v, err := base64.StdEncoding.DecodeString(to)
	if err != nil {
		return err
	}
	val.set(protoreflect.ValueOfBytes(v))
	return nil
}

func newMapValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "JSON object",
		stringFn: mapValueString,
		setFn:    mapValueSet,
	}
}

func mapValueString(val *Value) string {
	parent := val.getParent()
	fd := val.fieldDescriptor()
	if parent.Get(fd).Map().Len() == 0 {
		return ""
	}

	wrappedMsg := parent.Type().New()
	wrappedMsg.Set(fd, parent.Get(fd))
	wrappedBytes, err := protojson.Marshal(wrappedMsg.Interface())
	if err != nil {
		return ""
	}

	wrappedRaw := map[string]json.RawMessage{}
	if err := json.Unmarshal(wrappedBytes, &wrappedRaw); err != nil {
		return ""
	}
	return string(wrappedRaw[fd.JSONName()])
}

func mapValueSet(val *Value, to string) error {
	fd := val.fieldDescriptor()

	wrappedRaw := map[string]json.RawMessage{
		fd.JSONName(): []byte(to),
	}
	wrappedBytes, err := json.Marshal(wrappedRaw)
	if err != nil {
		return err
	}

	wrappedMsg := val.getParent().Type().New()
	if err := protojson.Unmarshal(wrappedBytes, wrappedMsg.Interface()); err != nil {
		return err
	}

	if !wrappedMsg.Has(fd) {
		val.set(wrappedMsg.NewField(fd))
		return nil
	}
	v := wrappedMsg.Get(fd)
	if !v.Map().IsValid() {
		val.set(wrappedMsg.NewField(fd))
		return nil
	}

	val.set(v)
	return nil
}

func newMessageValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "JSON object",
		stringFn: messageValueString,
		setFn:    messageValueSet,
	}
}

func messageValueString(val *Value) string {
	b, err := protojson.Marshal(val.GetValue().Message().Interface())
	if err != nil {
		return ""
	}
	str := string(b)
	if str == "{}" {
		return ""
	}
	return str
}

func messageValueSet(val *Value, to string) error {
	m := val.MessageType().New()
	if err := protojson.Unmarshal([]byte(to), m.Interface()); err != nil {
		return err
	}
	val.set(protoreflect.ValueOfMessage(m))
	return nil
}

func sliceValueString[T any](slc []T) string {
	if len(slc) == 0 {
		return ""
	}
	b, err := json.Marshal(slc)
	if err != nil {
		return ""
	}
	return string(b)
}

func newRepeatedBoolValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "bool (list)",
		stringFn: repeatedBoolValueString,
		setFn:    repeatedBoolValueSet,
		isBool:   true,
	}
}

func repeatedBoolValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]bool, li.Len())
	for i := range li.Len() {
		slc[i] = li.Get(i).Bool()
	}
	return sliceValueString(slc)
}

func repeatedBoolValueSet(val *Value, to string) error {
	v, err := strconv.ParseBool(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfBool(v))
	return nil
}

func newRepeatedEnumValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "string (list)",
		stringFn: repeatedEnumValueString,
		setFn:    repeatedEnumValueSet,
	}
}

func repeatedEnumValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]string, li.Len())
	for i := range li.Len() {
		v := li.Get(i).Enum()
		if desc := val.EnumDescriptor().Values().ByNumber(v); desc != nil {
			slc[i] = string(desc.Name())
			continue
		}
		slc[i] = strconv.FormatInt(int64(v), 10)
	}
	return sliceValueString(slc)
}

func repeatedEnumValueSet(val *Value, to string) error {
	desc := val.EnumDescriptor().Values().ByName(protoreflect.Name(to))
	if desc != nil {
		val.append(protoreflect.ValueOfEnum(desc.Number()))
		return nil
	}
	v, err := strconv.ParseInt(to, 10, 32)
	if err != nil {
		return fmt.Errorf("unknown enum value %q", to)
	}
	val.append(protoreflect.ValueOfEnum(protoreflect.EnumNumber(v)))
	return nil
}

func repeatedIntValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]int64, li.Len())
	for i := range li.Len() {
		slc[i] = li.Get(i).Int()
	}
	return sliceValueString(slc)
}

func newRepeatedInt32Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "int32 (list)",
		stringFn: repeatedIntValueString,
		setFn:    repeatedInt32ValueSet,
	}
}

func repeatedInt32ValueSet(val *Value, to string) error {
	v, err := parseInt32(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfInt32(v))
	return nil
}

func newRepeatedInt64Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "int64 (list)",
		stringFn: repeatedIntValueString,
		setFn:    repeatedInt64ValueSet,
	}
}

func repeatedInt64ValueSet(val *Value, to string) error {
	v, err := parseInt64(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfInt64(v))
	return nil
}

func repeatedUintValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]uint64, li.Len())
	for i := range li.Len() {
		slc[i] = li.Get(i).Uint()
	}
	return sliceValueString(slc)
}

func newRepeatedUint32Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "uint32 (list)",
		stringFn: repeatedUintValueString,
		setFn:    repeatedUint32ValueSet,
	}
}

func repeatedUint32ValueSet(val *Value, to string) error {
	v, err := parseUint32(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfUint32(v))
	return nil
}

func newRepeatedUint64Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "uint64 (list)",
		stringFn: repeatedUintValueString,
		setFn:    repeatedUint64ValueSet,
	}
}

func repeatedUint64ValueSet(val *Value, to string) error {
	v, err := parseUint64(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfUint64(v))
	return nil
}

func newRepeatedFloat32Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "float32 (list)",
		stringFn: repeatedFloatValueString,
		setFn:    repeatedFloat32ValueSet,
	}
}

func repeatedFloatValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]float64, li.Len())
	for i := range li.Len() {
		slc[i] = li.Get(i).Float()
	}
	return sliceValueString(slc)
}

func repeatedFloat32ValueSet(val *Value, to string) error {
	v, err := parseFloat32(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfFloat32(v))
	return nil
}

func newRepeatedFloat64Value(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "float64 (list)",
		stringFn: repeatedFloatValueString,
		setFn:    repeatedFloat64ValueSet,
	}
}

func repeatedFloat64ValueSet(val *Value, to string) error {
	v, err := parseFloat64(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfFloat64(v))
	return nil
}

func newRepeatedStringValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "string (list)",
		stringFn: repeatedStringValueString,
		setFn:    repeatedStringValueSet,
	}
}

func repeatedStringValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]string, li.Len())
	for i := range li.Len() {
		slc[i] = li.Get(i).String()
	}
	return sliceValueString(slc)
}

func repeatedStringValueSet(val *Value, to string) error {
	val.append(protoreflect.ValueOfString(to))
	return nil
}

func newRepeatedBytesValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "base64 (list)",
		stringFn: repeatedBytesValueString,
		setFn:    repeatedBytesValueSet,
	}
}

func repeatedBytesValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]string, li.Len())
	for i := range li.Len() {
		slc[i] = base64.StdEncoding.EncodeToString(li.Get(i).Bytes())
	}
	return sliceValueString(slc)
}

func repeatedBytesValueSet(val *Value, to string) error {
	v, err := base64.StdEncoding.DecodeString(to)
	if err != nil {
		return err
	}
	val.append(protoreflect.ValueOfBytes(v))
	return nil
}

func newRepeatedMessageValue(base protoreflect.Message, fields []protoreflect.FieldDescriptor) *Value {
	return &Value{
		base:     base,
		fields:   fields,
		typeName: "JSON object (list)",
		stringFn: repeatedMessageValueString,
		setFn:    repeatedMessageValueSet,
	}
}

func repeatedMessageValueString(val *Value) string {
	li := val.GetValue().List()
	slc := make([]json.RawMessage, li.Len())
	for i := range li.Len() {
		b, err := protojson.Marshal(li.Get(i).Message().Interface())
		if err != nil {
			return ""
		}
		slc[i] = b
	}
	return sliceValueString(slc)
}

func repeatedMessageValueSet(val *Value, to string) error {
	v := val.MessageType().New()
	if err := protojson.Unmarshal([]byte(to), v.Interface()); err != nil {
		return err
	}
	val.append(protoreflect.ValueOfMessage(v))
	return nil
}
