package packethandler

import (
	"context"
)

type _FuncPacketHandler struct {
	name     string
	beforeFn HandlerFn
	afterFn  HandlerFn
}

func NewFuncPacketHandler(name string, beforeFn HandlerFn, afterFn HandlerFn) (packHandler PacketHandlerI) {
	return &_FuncPacketHandler{
		name:     name,
		beforeFn: beforeFn,
		afterFn:  afterFn,
	}
}

func (packet *_FuncPacketHandler) Name() string {
	return packet.name
}

func (packet *_FuncPacketHandler) Description() string {
	return `将函数封装为packet`
}
func (packet *_FuncPacketHandler) Before(ctx context.Context, input []byte) (newCtx context.Context, out []byte, err error) {

	if packet.beforeFn == nil {
		return EmptyHandlerFn(ctx, input)
	}
	newCtx, out, err = packet.beforeFn(ctx, input)
	if err != nil {
		return ctx, nil, err
	}
	return newCtx, out, nil
}
func (packet *_FuncPacketHandler) After(ctx context.Context, input []byte) (newCtx context.Context, out []byte, err error) {
	if packet.afterFn == nil {
		return EmptyHandlerFn(ctx, input)
	}
	newCtx, out, err = packet.afterFn(ctx, input)
	if err != nil {
		return ctx, nil, err
	}
	return newCtx, out, nil
}

func (packet *_FuncPacketHandler) String() string {
	return ""
}
