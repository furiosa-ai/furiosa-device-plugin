package server

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func getRawJSON(i interface{}) []byte {
	if pb, ok := i.(proto.Message); ok {

		marshal, err := protojson.Marshal(pb)
		if err != nil {
			return nil
		}

		return marshal
	}
	return nil
}
