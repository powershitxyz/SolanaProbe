package parser

import (
	"errors"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/programs/vote"
)

type DecoderHandler func(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error)

type DecodeInsData struct {
	Plat string
	Key  string
	Data interface{}
}

var Handlers = map[string]DecoderHandler{
	"11111111111111111111111111111111":            handleSystemDecoder,
	"Vote111111111111111111111111111111111111111": handleVoteDecoder,
	"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA": handleTokenDecoder,
	Dexs["Raydium"]:      handleRadiumLiqV4,
	Dexs["JupiterV6"]:    handleJupiterV6,
	Dexs["Orca"]:         handleOrcaLiqV2,
	Dexs["Meteora"]:      handleMeteora,
	Dexs["MeteoraPools"]: handleMeteoraPoolsDex,
	Dexs["PumpFun"]:      handlePumpFun,
	Dexs["Raydiumv3"]:    handleRadiumV3,
	Dexs["RaydiumAmm"]:   handleRadiumAmm,
	Dexs["RaydiumCPMM"]:  handleRadiumCPMM,
	Dexs["Moonshot"]:     handleMoonshot,
	Dexs["OkxProxy"]:     handleOkxProxy,
}

func handlePumpFun(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(PumpFun)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleOkxProxy(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(OkxProxy)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}

var Dexs = map[string]string{
	"Raydium":      string(Raydium),      // Radium Liq V4
	"Raydiumv3":    string(Raydiumv3),    // Radium Liq V3
	"RaydiumCPMM":  string(RaydiumCPMM),  // Radium CPMM
	"RaydiumAmm":   string(RaydiumAmm),   // Raydium Liquidity Pool AMM
	"JupiterV6":    string(JupiterV6),    // Jupiter V6 Aggregator
	"Orca":         string(OrcaV2),       //
	"Meteora":      string(Meteora),      //
	"MeteoraPools": string(MeteoraPools), //
	"Fluxbeam":     string(FluxBeam),     //
	"PumpFun":      string(PumpFun),      //
	"Moonshot":     string(Moonshot),     //
	"OkxProxy":     string(OkxProxy),     //
}

func handleSystemDecoder(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	sysInstru, err := system.DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}

	typeName := system.InstructionIDToName(sysInstru.TypeID.Uint32())
	if typeName == "Transfer" {
		typeName = "STransfer"
	}
	return &DecodeInsData{
		Key:  typeName,
		Data: sysInstru,
	}, nil
}

func handleVoteDecoder(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	voteInstru, err := vote.DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}

	return &DecodeInsData{
		Key:  vote.ProgramName,
		Data: voteInstru,
	}, nil
}

func handleTokenDecoder(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	instru, err := token.DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}

	typeName := token.InstructionIDToName(instru.TypeID.Uint8())

	return &DecodeInsData{
		Key:  typeName,
		Data: instru,
	}, nil
}

func handleOrcaLiqV2(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(OrcaV2)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleMeteora(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(Meteora)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleMeteoraPoolsDex(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(MeteoraPools)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleFluxbeam(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(FluxBeam)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}

func handleRadiumLiqV4(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(Raydium)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)
	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleRadiumV3(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(Raydiumv3)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)
	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleRadiumCPMM(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(RaydiumCPMM)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)
	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleRadiumAmm(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(RaydiumAmm)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)
	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleMoonshot(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(Moonshot)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)
	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}

func handleJupiterV6(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(JupiterV6)
	if err != nil {
		return nil, err
	}
	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func handleOtherSwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	router, err := NewDexRouter(Other)
	if err != nil {
		return nil, err
	}

	result, err := router.UniCall(accounts, data, extra...)

	if err != nil || result == nil {
		return nil, err
	}
	if result == nil {
		log.Error("[handleOtherSwap] result nil")
		return nil, errors.New("handleOtherSwap nil")
	}
	return &DecodeInsData{
		Plat: result.DexName,
		Key:  result.TypeName,
		Data: result.Data,
	}, nil
}
func ParseDecode(program solana.PublicKey, accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {

	//spew.Dump(fmt.Sprintf("program: %s", program.String()))
	if handler, exists := Handlers[program.String()]; exists && handler != nil {
		return handler(accounts, data, extra...)
	}
	return nil, errors.New("no decoder found")
}
func ParseOtherDecode(program solana.PublicKey, accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*DecodeInsData, error) {
	if len(data) <= 0 {
		return nil, errors.New("data should not be empty")
	}
	//spew.Dump(fmt.Sprintf("program: %s", program.String()))
	return handleOtherSwap(accounts, data, extra...)

}
