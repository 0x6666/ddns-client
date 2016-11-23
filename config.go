package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/inimei/backup/log"
)

type cfgData struct {
	Key        string `toml:"key"`
	ServerHost string `toml:"server_host"`
	TickTime   int64  `toml:"tick_time"`
}

var Data cfgData

func CurDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

func init() {
	dir := CurDir()

	err := initialize(dir + "/ddns-client.toml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func initialize(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "ddns.toml"
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		cwd, _ := os.Getwd()
		log.Error("os.Stat fail, %s ,please ensure ddns.toml exist. ddns.toml path:%s, cwd:%s", err, configFilePath, cwd)
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error("read config file error, %s", err)
		return err
	}

	if err := toml.Unmarshal(buf, &Data); err != nil {
		log.Error("unmarshal config failed, %s", err)
		return err
	}

	if !strings.HasSuffix(Data.ServerHost, "/") && len(Data.ServerHost) != 0 {
		Data.ServerHost = Data.ServerHost + "/"
	}

	return nil
}
