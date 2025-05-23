package protobuf

import (
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoJson = protojson.MarshalOptions{
	Indent:            "  ",
	EmitDefaultValues: true,
	EmitUnpopulated:   true,
}

func MarshalIndent(msg any) ([]byte, error) {
	if protoMsg, ok := msg.(proto.Message); ok {
		return protoJson.Marshal(protoMsg)
	}
	if protoMsgs, ok := msg.([]proto.Message); ok {
		var jsonObjects []json.RawMessage
		for _, protoMsg := range protoMsgs {
			jsonObject, err := protoJson.Marshal(protoMsg)
			if err != nil {
				return nil, err
			}
			jsonObjects = append(jsonObjects, jsonObject)
		}

		return json.MarshalIndent(jsonObjects, "", "  ")
	}

	return json.MarshalIndent(msg, "", "  ")
}
