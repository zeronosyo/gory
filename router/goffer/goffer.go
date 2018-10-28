package goffer

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zeronosyo/gory/log"
)

func Register(router *gin.RouterGroup) {
	router.GET("/ping", ping)
}

func ping(c *gin.Context) {
	time.Sleep(5 * time.Second)
	log.AddLogMeta(c, "meta", "this_is_meta_data")
	log.AddLogArgs(c, "args1", "this_is_args1")
	log.AddLogArgs(c, "args2", 2)
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
