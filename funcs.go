package packethandler

import (
	"fmt"

	"github.com/suifengpiao14/funcs"
)

//GeneratePacketHandlerName 协助生成packetHandler名称，需要在 packetHander.Name()内部调用
func GeneratePacketHandlerName(packetHander PacketHandlerI) (name string) {
	funcName := funcs.GetCallFuncname(1)
	packetName, _ := funcs.SplitFullFuncName(funcName)
	structName := funcs.GetStructName(packetHander)
	return fmt.Sprintf("%s.%s", packetName, structName)
}
