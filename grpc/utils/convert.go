package utils

import (
	"encoding/json"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/encoding/protojson"
)

func StructToProtobufStruct(s interface{}) (*structpb.Struct, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	protobufStruct := &structpb.Struct{}
	err = protojson.Unmarshal(b, protobufStruct)
	if err != nil {
		return nil, err
	}
	return protobufStruct, nil
}

func ProtobufStructToStruct(s *structpb.Struct) (interface{}, error) {
	b, err := protojson.Marshal(s)
	if err != nil {
		return nil, err
	}

	var m interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
