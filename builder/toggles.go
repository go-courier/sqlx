package builder

import (
	"context"

	contextx "github.com/go-courier/x/context"
)

var (
	ToggleMultiTable    = "MultiTable"
	ToggleNeedAutoAlias = "NeedAlias"
	ToggleUseValues     = "UseValues"
)

type Toggles map[string]bool

func (toggles Toggles) Merge(next Toggles) Toggles {
	final := Toggles{}

	for k, v := range toggles {
		if v {
			final[k] = true
		}
	}

	for k, v := range next {
		if v {
			final[k] = true
		} else {
			delete(final, k)
		}
	}

	return final
}

func (toggles Toggles) Is(key string) bool {
	if v, ok := toggles[key]; ok {
		return v
	}
	return false
}

type contextKeyForToggles struct {
}

func ContextWithToggles(ctx context.Context, toggles Toggles) context.Context {
	return contextx.WithValue(ctx, contextKeyForToggles{}, TogglesFromContext(ctx).Merge(toggles))
}

func TogglesFromContext(ctx context.Context) Toggles {
	if ctx == nil {
		return Toggles{}
	}
	if toggles, ok := ctx.Value(contextKeyForToggles{}).(Toggles); ok {
		return toggles
	}
	return Toggles{}
}
