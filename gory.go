package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	goryLog "github.com/zeronosyo/gory/log"
	goryRouter "github.com/zeronosyo/gory/router"
)

const (
	EnvConfigFilename = "env.yml"
)

var (
	env    *Env
	logger *logrus.Logger
	router *gin.Engine
)

type Env struct {
	Addr            string `yaml:"addr"`
	Timeout         int32  `yaml:"timeout"`
	GracefulTimeout int32  `yaml:"graceful_timeout"`
}

func init() {
	logger = goryLog.InitLogger()
	env = LoadEnv()
	router = goryRouter.InitRouter()
}

/**********
 * Env *
 **********/

func LoadEnv() *Env {
	env := &Env{
		Addr:            "0.0.0.0:8000",
		Timeout:         20,
		GracefulTimeout: 30,
	}
	ymlFile, err := ioutil.ReadFile(EnvConfigFilename)
	if err != nil {
		logger.Fatalf("Load config env got error => %v", err)
	}
	if err := yaml.Unmarshal(ymlFile, env); err != nil {
		logger.Fatalf("Load config env got error => %v", err)
	}
	return env
}

/*********
 * Serve *
 *********/

func Serve(addr string, timeout int32, graceful_timeout int32) {
	w := logger.Writer()
	defer w.Close()
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    time.Duration(timeout) * time.Second,
		WriteTimeout:   time.Duration(timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.New(w, "", 0),
	}
	go func(srv *http.Server) {
		logger.Println("Starting Server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}(srv)
	// catch os interrupt signal
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Println("Shutdown Server...")
	// gracefull timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(graceful_timeout)*time.Second,
	)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown: ", err)
	}
	logger.Println("Server exit")
}

func main() {
	Serve(env.Addr, env.Timeout, env.GracefulTimeout)
}
