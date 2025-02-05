package config

import (
	"fmt"
	"testing"
)

func TestConfigInit(t *testing.T) {
	rpcList := systemConfig.Chain.GetRpc()
	rpcMapList := systemConfig.Chain.GetRpcMapper()
	fmt.Println(rpcList, rpcMapList)

	i := systemConfig.Chain.RpcMap[rpcList[0]]
	fmt.Println(i)
}
