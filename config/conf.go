package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	system "github.com/powershitxyz/SolanaProbe/sys"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database    DatabaseConfig `yaml:"database"`
	Redis       RedisConfig    `yaml:"redis"`
	Chain       ChainConfig    `yaml:"chain"`
	Log         LogConfig      `yaml:"log"`
	AllStart    int            `yaml:"allStart"`
	FeeAccounts FeeAccounts    `yaml:"feeAccounts"`
}

// DatabaseConfig holds the database connection parameters.
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"TimeZone"`
}

// RedisConfig holds the Redis connection parameters.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

// ChainConfig holds the Solana chain RPC endpoints.
type ChainConfig struct {
	Name         string   `yaml:"name"`
	WsRpc        string   `yaml:"wsRpc"`
	QueryRpc     []string `yaml:"queryRpc"`
	SlotParallel int      `yaml:"slotParallel"`
	TxDetal      int      `yaml:"txDetal"`
	RangeRound   int      `yaml:"rangeRound"`
	Rpcs         []RpcMapper
	RpcMap       map[string]int
}

// LogConfig holds the logging directory and file name.
type LogConfig struct {
	Path string `yaml:"path"`
	Name string `yaml:"name"`
}

type RpcMapper struct {
	Rpc   string
	Quote int
}
type FeeAccounts struct {
	SOL  string `yaml:"SOL"`
	WSOL string `yaml:"WSOL"`
	USDC string `yaml:"USDC"`
	USDT string `yaml:"USDT"`
}

func (f *FeeAccounts) GetFeeAccount() map[string]string {
	feeAccount := make(map[string]string)
	feeAccount[f.SOL] = "SOL"
	wsolSplit := strings.Split(f.WSOL, ":")
	feeAccount[wsolSplit[1]] = wsolSplit[0]
	usdcSplit := strings.Split(f.USDC, ":")
	feeAccount[usdcSplit[1]] = usdcSplit[0]
	usdtSplit := strings.Split(f.USDT, ":")
	feeAccount[usdtSplit[1]] = usdtSplit[0]
	return feeAccount
}

func (t *ChainConfig) initRpc() {
	if len(t.QueryRpc) == 0 {
		log.Fatal("error rpc config")
	}

	t.RpcMap = make(map[string]int)
	for _, r := range t.QueryRpc {
		v := strings.Split(r, "||")
		num := 0
		if len(v) == 2 {
			numStr := v[1]
			var err error
			num, err = strconv.Atoi(numStr)
			if err != nil {
				log.Fatal("error rpc inner format with Quote:", err)
			}
		}

		t.Rpcs = append(t.Rpcs, RpcMapper{
			Rpc:   v[0],
			Quote: num,
		})
		t.RpcMap[v[0]] = num
	}
}

func (t ChainConfig) GetRpc() []string {
	r := make([]string, 0)
	for _, v := range t.Rpcs {
		r = append(r, v.Rpc)
	}
	return r
}

func (t ChainConfig) GetRpcMapper() []RpcMapper {
	return t.Rpcs
}

func (t *ChainConfig) GetSlotParallel() int {
	if t.SlotParallel > 0 {
		return t.SlotParallel
	}
	return 1
}

func (t *ChainConfig) GetTxDelay() int {
	if t.TxDetal > 0 {
		return t.TxDetal
	}
	return 0
}

var systemConfig = &Config{}

func GetConfig() Config {
	return *systemConfig
}

func findProjectRoot(currentDir, rootIndicator string) (string, error) {
	if _, err := os.Stat(filepath.Join(currentDir, rootIndicator)); err == nil {
		return currentDir, nil
	}
	parentDir := filepath.Dir(currentDir)
	if currentDir == parentDir {
		return "", os.ErrNotExist
	}
	return findProjectRoot(parentDir, rootIndicator)
}

func init() {
	winFlag := false
	switch os := runtime.GOOS; os {
	case "windows":
		winFlag = true
		fmt.Println("当前系统是 Windows")
	case "linux":
		fmt.Println("当前系统是 Linux")
	default:
		fmt.Printf("当前系统是 %s\n", os)
	}
	var confFilePath string

	if configFilePathFromEnv := os.Getenv("DALINK_GO_CONFIG_PATH"); configFilePathFromEnv != "" {
		confFilePath = configFilePathFromEnv
	} else {
		_, filename, _, _ := runtime.Caller(0)
		testDir := filepath.Dir(filename)
		confFilePath, _ = findProjectRoot(testDir, "__mark__")
		if len(confFilePath) > 0 {
			confFilePath += "/config/dev.yml"
		}
	}
	if len(confFilePath) == 0 {
		log.Fatal("系统根初始化失败")
	}

	viper.SetConfigFile(confFilePath)

	viper.SetConfigType("yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("无法读取配置文件：%s", err)
	}

	err = viper.Unmarshal(&systemConfig)
	if err != nil {
		log.Fatalf("无法解析配置：%s", err)
	}
	systemConfig.Chain.initRpc()
	if winFlag {
		system.LogFile, err = os.OpenFile(systemConfig.Log.Name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		system.InitLogger("")
	} else {
		system.LogFile, err = os.OpenFile(systemConfig.Log.Path+systemConfig.Log.Name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		system.InitLogger(systemConfig.Log.Path)
	}

	_ = godotenv.Load()
	system.Logger.Printf("initing default chain config %+v", systemConfig)
	system.FeeAddrs = systemConfig.FeeAccounts.GetFeeAccount()
}
