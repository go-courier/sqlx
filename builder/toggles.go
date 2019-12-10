package builder

import (
	"context"
)

var (
	keyForToggles = "$$builder.toggles"
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

func ContextWithToggles(ctx context.Context, toggles Toggles) context.Context {
	return context.WithValue(ctx, keyForToggles, TogglesFromContext(ctx).Merge(toggles))
}

func TogglesFromContext(ctx context.Context) Toggles {
	if ctx == nil {
		return Toggles{}
	}
	if toggles, ok := ctx.Value(keyForToggles).(Toggles); ok {
		return toggles
	}
	return Toggles{}
}
