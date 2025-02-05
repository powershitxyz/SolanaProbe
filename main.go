package main

import (
	_ "net/http/pprof"
	"sync"

	"github.com/powershitxyz/SolanaProbe/rpc"
	"github.com/powershitxyz/SolanaProbe/sys"
)

var logger = sys.Logger
var wg sync.WaitGroup

func main() {
	rpc.InitEssential()
}
