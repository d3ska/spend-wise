package auth

import (
	"context"
	"testing"

	"backend/model"
)

func TestWithUserIDAndUserIDFromContext(t *testing.T) {
	tests := map[string]struct {
		setup func(context.Context) context.Context
		want  model.UserID
		ok    bool
	}{
		"store and retrieve user ID": {
			setup: func(ctx context.Context) context.Context {
				return WithUserID(ctx, 42)
			},
			want: 42,
			ok:   true,
		},
		"missing user ID": {
			setup: func(ctx context.Context) context.Context {
				return ctx
			},
			want: 0,
			ok:   false,
		},
		"nested contexts - inner overrides outer": {
			setup: func(ctx context.Context) context.Context {
				ctx = WithUserID(ctx, 10)
				ctx = WithUserID(ctx, 20)
				return ctx
			},
			want: 20,
			ok:   true,
		},
		"zero user ID is valid": {
			setup: func(ctx context.Context) context.Context {
				return WithUserID(ctx, 0)
			},
			want: 0,
			ok:   true,
		},
		"large user ID": {
			setup: func(ctx context.Context) context.Context {
				return WithUserID(ctx, 9223372036854775807) // max int64
			},
			want: 9223372036854775807,
			ok:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx = tt.setup(ctx)

			got, ok := UserIDFromContext(ctx)
			if ok != tt.ok {
				t.Errorf("UserIDFromContext() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("UserIDFromContext() got %v, want %v", got, tt.want)
			}
		})
	}
}
