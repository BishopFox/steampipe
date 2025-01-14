package controlhooks

import "context"

var NullHooks = &NullControlHook{}

type NullControlHook struct{}

func (*NullControlHook) OnStart(context.Context, *ControlProgress)        {}
func (*NullControlHook) OnControlEvent(context.Context, *ControlProgress) {}
func (*NullControlHook) OnDone(context.Context, *ControlProgress)         {}
