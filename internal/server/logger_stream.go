package server

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type wrappedServerStream struct {
	grpc.ServerStream
	ctx    context.Context
	logger *zerolog.Logger
	info   *grpc.StreamServerInfo
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	timestamp := time.Now()

	err := w.ServerStream.SendMsg(m)
	defer func() {
		var event *zerolog.Event
		if err != nil {
			event = getNewErrorEventStreamLogger(w.logger, timestamp, m, w.info, err)
		} else {
			event = getNewDebugEventStreamLogger(w.logger, timestamp, m, w.info)
		}

		event.Msg("grpc middleware event stream send logging")
	}()

	return err
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	timestamp := time.Now()

	err := w.ServerStream.RecvMsg(m)
	defer func() {
		var event *zerolog.Event
		if err != nil {
			event = getNewErrorEventStreamLogger(w.logger, timestamp, m, w.info, err)
		} else {
			event = getNewDebugEventStreamLogger(w.logger, timestamp, m, w.info)
		}

		event.Msg("grpc middleware event stream recv logging")
	}()

	return err
}

func NewGrpcStreamLogger(ctx context.Context) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger := zerolog.Ctx(ctx)
		streamCtx := logger.WithContext(ss.Context())

		wss := &wrappedServerStream{
			ServerStream: ss,
			ctx:          streamCtx,
			logger:       logger,
			info:         info,
		}

		return handler(srv, wss)
	}
}

func getNewErrorEventStreamLogger(logger *zerolog.Logger, time time.Time, m interface{}, info *grpc.StreamServerInfo, err error) *zerolog.Event {
	statusErr := status.Convert(err)
	event := logger.Err(err).Time(zerolog.TimestampFieldName, time).Str("method", info.FullMethod).Str("error_code", statusErr.Code().String()).Str("msg", statusErr.Message()).Interface("details", statusErr.Details())
	if raw := getRawJSON(m); raw != nil {
		event = event.RawJSON("message", raw)
	}

	return event
}

func getNewDebugEventStreamLogger(logger *zerolog.Logger, time time.Time, m interface{}, info *grpc.StreamServerInfo) *zerolog.Event {
	event := logger.Debug().Time(zerolog.TimestampFieldName, time).Str("method", info.FullMethod)
	if raw := getRawJSON(m); raw != nil {
		event = event.RawJSON("message", raw)
	}

	return event
}
