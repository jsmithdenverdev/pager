package codec

import (
	"context"

	"github.com/jsmithdenverdev/pager/pkg/valid"
)

type EncoderOption[TOut any] func(*TOut)

type Encoder[TIn, TOut any] interface {
	Encode(ctx context.Context, in TIn, opts ...EncoderOption[TOut]) (TOut, error)
}

type Decoder[TIn any, TOut valid.Validator] interface {
	Decode(ctx context.Context, in TIn) (TOut, error)
}

type EncoderFunc[TIn, TOut any] func(ctx context.Context, in TIn, opts ...EncoderOption[TOut]) (TOut, error)

func (fn EncoderFunc[TIn, TOut]) Encode(ctx context.Context, in TIn, opts ...EncoderOption[TOut]) (TOut, error) {
	return fn(ctx, in, opts...)
}

type DecoderFunc[TIn any, TOut valid.Validator] func(ctx context.Context, in TIn) (TOut, error)

func (fn DecoderFunc[TIn, TOut]) Decode(ctx context.Context, in TIn) (TOut, error) {
	return fn(ctx, in)
}
