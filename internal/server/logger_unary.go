package server

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func NewGrpcUnaryLogger(ctx context.Context) grpc.UnaryServerInterceptor {
	return func(reqCtx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var event *zerolog.Event
		timestamp := time.Now()

		logger := zerolog.Ctx(ctx)
		reqCtxWithLogger := logger.WithContext(reqCtx)

		resp, err := handler(reqCtxWithLogger, req)
		if err != nil {
			//Note: error logging is enabled by default
			event = getNewErrorEventUnaryLogger(logger, timestamp, req, info, err)
		} else {
			//TODO(@bg): enable if debugging mode is enabled
			event = getNewDebugEventUnaryLogger(logger, timestamp, req, info, resp)
		}

		event.Msg("grpc middleware event unary logging")

		return resp, err
	}
}

func getNewErrorEventUnaryLogger(logger *zerolog.Logger, time time.Time, req interface{}, info *grpc.UnaryServerInfo, err error) *zerolog.Event {
	statusErr := status.Convert(err)
	event := logger.Err(err).Time(zerolog.TimestampFieldName, time).Str("method", info.FullMethod).Str("error_code", statusErr.Code().String()).Str("msg", statusErr.Message()).Interface("details", statusErr.Details())

	if raw := getRawJSON(req); raw != nil {
		event = event.RawJSON("request", raw)
	}

	return event
}

func getNewDebugEventUnaryLogger(logger *zerolog.Logger, time time.Time, req interface{}, info *grpc.UnaryServerInfo, resp interface{}) *zerolog.Event {
	event := logger.Debug().Time(zerolog.TimestampFieldName, time).Str("method", info.FullMethod)

	if raw := getRawJSON(req); raw != nil {
		event = event.RawJSON("request", raw)
	}

	if raw := getRawJSON(resp); raw != nil {
		event = event.RawJSON("response", raw)
	}

	return event
}
