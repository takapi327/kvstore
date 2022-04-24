package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	abciclient "github.com/tendermint/tendermint/abci/client"
	cfg "github.com/tendermint/tendermint/config"
	tmlog "github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/types"
)

var homeDir string

func init() {
	flag.StringVar(&homeDir, "tm-home", "",  "Path to the tendermint config directory (if empty, uses $HOME/.tendermint)")
}

func main()  {
	flag.Parse()
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/.tendermint")
	}

	// Tendermint Coreの設定ファイルを読み込む (Start)
	config := cfg.DefaultValidatorConfig()

	config.SetRoot(homeDir)

	viper.SetConfigFile(fmt.Sprintf("%s/%s", homeDir, "config/config.toml"))
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Reading config: %v", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("Decoding config: %v", err)
	}
	if err := config.ValidateBasic(); err != nil {
		log.Fatalf("Invalid configuration data: %v", err)
	}
	gf, err := types.GenesisDocFromFile(config.GenesisFile())
	if err != nil {
		log.Fatalf("Loading genesis document: %v", err)
	}
	// Tendermint Coreの設定ファイルを読み込む (End)

	// データベースハンドルを作成し、それを使ってABCIアプリケーションを構築 (Start)
	dbPath := filepath.Join(homeDir, "badger")
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Closing database: %v", err)
		}
	}()
	app := NewKVStoreApplication(db)
	acc := abciclient.NewLocalCreator(app)
	// データベースハンドルを作成し、それを使ってABCIアプリケーションを構築 (End)

	// ロガーを構築
	logger := tmlog.MustNewDefaultLogger(tmlog.LogFormatPlain, tmlog.LogLevelInfo, false)

	// 設定、ロガー、アプリケーションへのハンドル、そしてgenesisファイルを渡すことでノードを構築 (Start)
	node, err := nm.New(config, logger, acc, gf)
	if err != nil {
		log.Fatalf("Creating node: %v", err)
	}
	// 設定、ロガー、アプリケーションへのハンドル、そしてgenesisファイルを渡すことでノードを構築 (End)

	// ノードを起動
	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	// このロジックにより、プログラムはSIGTERMをキャッチすることがでるようになる。
	// これは、オペレータがプログラムを終了させようとしたときに、ノードが優雅にシャットダウンできるようにするためのもの
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<- c
}
