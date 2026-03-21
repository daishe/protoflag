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

package genreflect

import (
	protoflagtestv1 "github.com/daishe/protoflag/internal/protoflagtest/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func md[T proto.Message]() protoreflect.MessageDescriptor {
	var zero T
	return zero.ProtoReflect().Descriptor()
}

var (
	Top_Value_field  = md[*protoflagtestv1.Top]().Fields().ByName("value")
	Top_Middle_field = md[*protoflagtestv1.Top]().Fields().ByName("middle")
)

var (
	Middle_Value_field  = md[*protoflagtestv1.Middle]().Fields().ByName("value")
	Middle_Bottom_field = md[*protoflagtestv1.Middle]().Fields().ByName("bottom")
)

var (
	Bottom_Value_field             = md[*protoflagtestv1.Bottom]().Fields().ByName("value")
	Bottom_ValueWithLongName_field = md[*protoflagtestv1.Bottom]().Fields().ByName("value_with_long_name")
)
