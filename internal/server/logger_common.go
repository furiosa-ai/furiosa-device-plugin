package server

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func getRawJSON(i interface{}) []byte {
	switch msg := i.(type) {
	case proto.Message:
		marshaled, err := protojson.Marshal(msg)
		if err != nil {
			return nil
		}

		return marshaled

	default:
		marshaled, err := json.Marshal(msg)
		if err != nil {
			return nil
		}

		return marshaled
	}
}
