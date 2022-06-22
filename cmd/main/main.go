package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/oklog/oklog/pkg/group"
	"github.com/spf13/viper"

	_ "github.com/weixiaodong/tanzhi/api"
	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/config"
	"github.com/weixiaodong/tanzhi/crontab"
	"github.com/weixiaodong/tanzhi/internal/job/scheduler"
	"github.com/weixiaodong/tanzhi/internal/store/db"
	httpserver "github.com/weixiaodong/tanzhi/transport/http"
)

var (
	env = flag.String("env", "config", "config file name")
)

func readConfig() *viper.Viper {
	v := viper.New()

	flag.Parse()
	configpath := "./config"
	// 如果env使用的是绝对路径，则configpath为路径，env为文件名
	if filepath.IsAbs(*env) {
		configpath, *env = filepath.Split(*env)
		// 去除.toml 文件后缀
		*env = strings.TrimSuffix(*env, ".toml")
	}

	v.SetConfigName(*env)
	v.AddConfigPath(configpath)

	// 找到并读取配置文件并且 处理错误读取配置文件
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	return v
}

func Init() {
	// 读取配置配置
	v := readConfig()

	// 初始化配置
	config.Init(v)

	// 初始化db
	db.Init()

	// 启动crontab定时任务
	crontab.Start()

	// 启动任务调度
	scheduler.Start()
}

func run() {
	var (
		g            group.Group
		err          error
		httpListener net.Listener
	)
	c := config.GConfig

	httpListener, err = net.Listen("tcp", c.Transport.HTTP.Addr)
	if err != nil {
		panic(fmt.Errorf("Start http server error: %s", err.Error()))
	}

	ctx := context.TODO()
	ctx = log.WithModule(ctx, "main")
	log.Info(ctx, "start_http_server", "addr", c.Transport.HTTP.Addr)

	g.Add(func() error {
		serverHandler := httpserver.NewHTTPHandler()
		httpServer := &http.Server{
			Handler:           serverHandler,
			ReadTimeout:       time.Duration(c.Transport.HTTP.ReadTimeout*1000) * time.Millisecond,
			ReadHeaderTimeout: time.Duration(c.Transport.HTTP.ReadHeaderTimeout*1000) * time.Millisecond,
			WriteTimeout:      time.Duration(c.Transport.HTTP.WriteTimeout*1000) * time.Millisecond,
			IdleTimeout:       time.Duration(c.Transport.HTTP.IdleTimeout*1000) * time.Millisecond,
		}

		return httpServer.Serve(httpListener)
	}, func(error) {
		httpListener.Close()
	})

	cancelInterrupt := make(chan struct{})
	// signal shutdown
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s shutdown", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})

	err = g.Run()

	log.Error(ctx, "exit_server", "err", err)

}

func main() {
	Init()

	run()
}
