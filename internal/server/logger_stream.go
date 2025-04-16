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

	ctx       context.Context
	logger    *zerolog.Logger
	info      *grpc.StreamServerInfo
	debugMode bool
}

func (wss *wrappedServerStream) Context() context.Context {
	return wss.ctx
}

func (wss *wrappedServerStream) SendMsg(m interface{}) error {
	timestamp := time.Now()

	err := wss.ServerStream.SendMsg(m)
	if err != nil {
		event := getNewErrorEventStreamLogger(wss.logger, timestamp, m, wss.info, err)
		event.Msg("grpc middleware event stream send error logging")
	}

	if wss.debugMode {
		event := getNewDebugEventStreamLogger(wss.logger, timestamp, m, wss.info)
		event.Msg("grpc middleware event stream send debug logging")
	}

	return err
}

func (wss *wrappedServerStream) RecvMsg(m interface{}) error {
	timestamp := time.Now()

	err := wss.ServerStream.SendMsg(m)
	if err != nil {
		event := getNewErrorEventStreamLogger(wss.logger, timestamp, m, wss.info, err)
		event.Msg("grpc middleware event stream recv error logging")
	}

	if wss.debugMode {
		event := getNewDebugEventStreamLogger(wss.logger, timestamp, m, wss.info)
		event.Msg("grpc middleware event stream recv debug logging")
	}

	return err
}

func NewGrpcLoggerStreamInterceptor(ctx context.Context, debugMode bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger := zerolog.Ctx(ctx)

		wss := &wrappedServerStream{
			ServerStream: ss,
			ctx:          logger.WithContext(ss.Context()),
			logger:       logger,
			info:         info,
			debugMode:    debugMode,
		}

		return handler(srv, wss)
	}
}

func getNewErrorEventStreamLogger(logger *zerolog.Logger, time time.Time, m interface{}, info *grpc.StreamServerInfo, err error) *zerolog.Event {
	statusErr := status.Convert(err)
	event := logger.Err(err).Time(zerolog.TimestampFieldName, time).Str("method", info.FullMethod).Str("error_code", statusErr.Code().String()).Str("msg", statusErr.Message()).Interface("details", statusErr.Details())
	if raw := getRawJSON(m); raw != nil {
		event = event.RawJSON("payload", raw)
	}

	return event
}

func getNewDebugEventStreamLogger(logger *zerolog.Logger, time time.Time, m interface{}, info *grpc.StreamServerInfo) *zerolog.Event {
	event := logger.Debug().Time(zerolog.TimestampFieldName, time).Str("method", info.FullMethod)
	if raw := getRawJSON(m); raw != nil {
		event = event.RawJSON("payload", raw)
	}

	return event
}
