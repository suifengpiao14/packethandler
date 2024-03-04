package packethandler

import (
	"bytes"
	"context"
	"fmt"

	"github.com/suifengpiao14/logchan/v2"
)

const (
	PacketLogName LogName = "packetLog"
)

type LogName string

func (l LogName) String() string {
	return string(l)
}

const (
	HandlerLog_Type_SetContext = "setContext"
	HandlerLog_Type_Before     = "before"
	HandlerLog_Type_After      = "after"
)

type HandlerLog struct {
	BeforeCtx context.Context
	AfterCtx  context.Context
	PackName  string
	Order     int
	Type      string // before |after
	Input     []byte `json:"input"`
	Output    []byte `json:"output"`
	Serialize string `json:"serialize"`
	Err       error
}

type PacketLog struct {
	HandlerLogs []HandlerLog `json:"handlerLogs"`
	logchan.EmptyLogInfo
}

func (w PacketLog) GetName() (logName logchan.LogName) {
	return PacketLogName
}

func (w PacketLog) Error() (err error) {
	for _, packHandlerLog := range w.HandlerLogs {
		if packHandlerLog.Err != nil {
			return packHandlerLog.Err
		}
	}
	return nil
}

const ColorRed = "\033[0;31m"
const ColorNone = "\033[0m"

func DefaultPrintStreamLog(logInfo logchan.LogInforInterface, typeName logchan.LogName, err error) {
	if typeName != PacketLogName {
		return
	}
	streamLog, ok := logInfo.(*PacketLog)
	if !ok {
		return
	}
	processSessionID := logchan.GetSessionID(logInfo)
	// 把没有出错的步骤日志输出，方便定位问题
	// if err != nil {
	// 	fmt.Fprintf(logchan.LogWriter, "processSessionID:%s|loginInfo:%s|error:%s\n", processSessionID, streamLog.GetName(), err.Error())
	// 	return
	// }
	fmt.Fprintf(logchan.LogWriter, "---------------------------begin------------------------\n")
	for i, handlerLog := range streamLog.HandlerLogs {
		errStr := ""
		if handlerLog.Err != nil {
			errStr = handlerLog.Err.Error()
		}

		if len(handlerLog.Input) > 0 && bytes.Equal(handlerLog.Input, handlerLog.Output) && handlerLog.BeforeCtx == handlerLog.AfterCtx { // 输入输出完全一致,上下文变量地址没改变,说明没有对数据处理，只是因为流程流过而已，则不输出日志
			continue
		}

		fmt.Fprintf(logchan.LogWriter,
			"processSessionID:%s|name:%s|serialNumber:%d|type:%s|input:%s|curryData:%s|err:%s%s%s\n",
			processSessionID,
			handlerLog.PackName,
			i,
			handlerLog.Type,
			string(handlerLog.Input),
			//string(handlerLog.Output),
			handlerLog.Serialize,
			ColorRed,
			errStr,
			ColorNone,
		)
	}
	fmt.Fprintf(logchan.LogWriter, "---------------------------end------------------------\n")
}
