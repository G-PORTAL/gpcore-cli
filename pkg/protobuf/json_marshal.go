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
	switch msg.(type) {
	case proto.Message:
		return protoJson.Marshal(msg.(proto.Message))
	case []proto.Message:
		var jsonObjects []json.RawMessage
		for _, protoMsg := range msg.([]proto.Message) {
			jsonObject, err := protoJson.Marshal(protoMsg)
			if err != nil {
				return nil, err
			}
			jsonObjects = append(jsonObjects, jsonObject)
		}

		return json.MarshalIndent(jsonObjects, "", "  ")
	default:
		return json.MarshalIndent(msg, "", "  ")
	}
}
