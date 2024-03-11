package packethandler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/logchan/v2"
)

type Flow []string

func ToFlow(s string) (flow Flow) {
	flow = Flow(strings.Split(s, ","))
	flow.DropEmpty()
	return flow
}

//DropEmpty 过滤空值
func (fs *Flow) DropEmpty() {
	flows := make(Flow, 0)
	for _, f := range *fs {
		s := strings.TrimSpace(f)
		if s == "" {
			continue
		}
		flows = append(flows, f)
	}
	*fs = flows
}

func (fs *Flow) String() (s string) {
	s = strings.Join(*fs, ",")
	return s
}

var ERROR_EMPTY_FUNC = errors.New("empty func")

// 定回调函数指针的类型
type HandlerFn func(ctx context.Context, input []byte) (newCtx context.Context, out []byte, err error)

func EmptyHandlerFn(ctx context.Context, input []byte) (newCtx context.Context, out []byte, err error) {
	return ctx, input, ERROR_EMPTY_FUNC
}

type PacketHandlerI interface {
	Name() string
	Description() string
	Before(ctx context.Context, input []byte) (newCtx context.Context, out []byte, err error)
	After(ctx context.Context, input []byte) (newCtx context.Context, out []byte, err error)
	String() string
}

// JsonString 用户实现PacketHandlerI.String()
func JsonString(packet PacketHandlerI) string {
	b, _ := json.Marshal(packet)
	s := string(b)
	return s
}

type PacketHandlers []PacketHandlerI

func NewPacketHandlers(packetHandlerIs ...PacketHandlerI) (packetHandlers PacketHandlers) {
	packetHandlers = make(PacketHandlers, 0)
	packetHandlers.Append(packetHandlerIs...)
	return packetHandlers
}

func (ps *PacketHandlers) Append(packetHandlers ...PacketHandlerI) {
	if *ps == nil {
		*ps = make(PacketHandlers, 0)
	}
	*ps = append(*ps, packetHandlers...)
}

// GetByName 通过名称获取子集合
func (ps *PacketHandlers) GetNames() (names []string) {
	names = make([]string, 0)
	for _, p := range *ps {
		names = append(names, p.Name())
	}
	return names
}

// GetByName 通过名称获取子集合
func (ps *PacketHandlers) GetByName(names ...string) (packetHandlers PacketHandlers, err error) {
	packetHandlers = make(PacketHandlers, 0)
	for _, name := range names {
		exists := false
		for _, p := range *ps {
			if strings.EqualFold(p.Name(), name) {
				packetHandlers.Append(p)
				exists = true
				break
			}
		}
		if !exists {
			err = errors.Errorf("not found packet handler named:%s", name)
			return nil, err
		}
	}

	return packetHandlers, nil
}

func (ps *PacketHandlers) InsertBefore(index int, packetHandlers ...PacketHandlerI) {
	if index <= 0 || index > len(*ps)-1 { // 找不到模板包位置，或者找到第一个，直接插入开头
		tmp := *ps
		*ps = make(PacketHandlers, 0)
		ps.Append(packetHandlers...)
		ps.Append(tmp...)
		return
	}
	before, after := (*ps)[0:index], (*ps)[index:]
	*ps = make(PacketHandlers, 0) // 此处必须重新申请，否则操作会覆盖原有地址
	ps.Append(before...)
	ps.Append(packetHandlers...)
	ps.Append(after...)

}

func (ps *PacketHandlers) InsertAfter(index int, packetHandlers ...PacketHandlerI) {
	if index < 0 || index+1 >= len(*ps) { // 找不到模板包位置,或者目标本就是最后一个，直接在结尾追加
		ps.Append(packetHandlers...)
		return
	}
	before, after := (*ps)[0:index+1], (*ps)[index+1:]
	*ps = make(PacketHandlers, 0) // 此处必须重新申请，否则操作会覆盖原有地址
	ps.Append(before...)
	ps.Append(packetHandlers...)
	ps.Append(after...)
}

func (ps *PacketHandlers) Delete(index int) {
	if index < 0 || len(*ps)-1 < index { // 越界不操作
		return
	}
	if index == len(*ps)-1 { // 需要删除的，在最后一个，直接截断
		*ps = (*ps)[:index]
		return
	}
	before, after := (*ps)[0:index], (*ps)[index+1:]
	*ps = make(PacketHandlers, 0) // 此处必须重新申请，否则操作会覆盖原有地址
	ps.Append(before...)
	ps.Append(after...)
}

func (ps *PacketHandlers) Replace(index int, packetHandlers ...PacketHandlerI) {
	if index < 0 || len(*ps)-1 < index { // 找不到模板包位置,不删除
		return
	}
	before, after := (*ps)[0:index], (*ps)[index+1:]
	*ps = make(PacketHandlers, 0) // 此处必须重新申请，否则操作会覆盖原有地址
	ps.Append(before...)
	ps.Append(packetHandlers...)
	ps.Append(after...)
}

func (ps *PacketHandlers) Index(name string) (indexs []int) {
	indexs = make([]int, 0)
	for i, packet := range *ps {
		if packet.Name() == name {
			indexs = append(indexs, i)
		}
	}
	return indexs
}
func (ps *PacketHandlers) IndexFirst(name string) (index int) {
	for i, packet := range *ps {
		if packet.Name() == name {
			return i
		}
	}
	return -1
}

func (ps *PacketHandlers) IndexLast(name string) (index int) {
	for i := len(*ps) - 1; i < 0; i++ {
		if (*ps)[i].Name() == name {
			return i
		}
	}
	return -1
}

func (ps PacketHandlers) Run(ctx context.Context, input []byte) (out []byte, err error) {
	data := input
	l := len(ps)
	streamLog := PacketLog{
		HandlerLogs: make([]HandlerLog, 0),
	}
	defer func() {
		streamLog.SetContext(ctx)
		logchan.SendLogInfo(&streamLog)
	}()
	for i := 0; i < l; i++ { // 先执行最后的before，直到最早的before
		pack := ps[i]
		handlerLog := HandlerLog{
			BeforeCtx: ctx,
			Input:     data,
			PackName:  pack.Name(),
			Type:      HandlerLog_Type_Before,
			Serialize: pack.String(),
		}
		ctx, data, err = pack.Before(ctx, data)
		if errors.Is(err, ERROR_EMPTY_FUNC) { // 这个错误标记是空函数，可以不计日志
			err = nil
			continue
		}
		handlerLog.Err = err
		handlerLog.AfterCtx = ctx
		handlerLog.Output = data
		streamLog.HandlerLogs = append(streamLog.HandlerLogs, handlerLog)
		if err != nil {
			return nil, err
		}
	}

	for i := l - 1; i > -1; i-- { // 先执行最后的after，直到最早的after
		pack := ps[i]
		handlerLog := HandlerLog{
			BeforeCtx: ctx,
			Input:     data,
			PackName:  pack.Name(),
			Type:      HandlerLog_Type_After,
			Serialize: pack.String(),
		}
		handlerLog.Input = data
		ctx, data, err = pack.After(ctx, data)
		if errors.Is(err, ERROR_EMPTY_FUNC) { // 这个错误标记是空函数，可以不计日志
			err = nil
			continue
		}
		handlerLog.Err = err
		handlerLog.AfterCtx = ctx
		handlerLog.Output = data
		streamLog.HandlerLogs = append(streamLog.HandlerLogs, handlerLog)
		if err != nil {
			return nil, err
		}
	}
	return data, err
}
