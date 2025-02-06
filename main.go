package main

import (
	_ "net/http/pprof"

	"github.com/powershitxyz/SolanaProbe/rpc"
)

func main() {
	go rpc.InitEssential()
	select {}
}
