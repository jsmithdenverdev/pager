package codec

import (
	"context"

	"github.com/jsmithdenverdev/pager/pkg/valid"
)

type EncoderOption[TOut any] func(*TOut)

type Encoder[TIn, TOut any] func(ctx context.Context, in TIn, opts ...EncoderOption[TOut]) (TOut, error)

type Decoder[TIn any, TOut valid.Validator] func(ctx context.Context, in TIn) (TOut, error)
