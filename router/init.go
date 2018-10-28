package router

import (
	"github.com/gin-gonic/gin"

	"github.com/zeronosyo/gory/log"
	"github.com/zeronosyo/gory/router/goffer"
)

func InitRouter() *gin.Engine {
	router := gin.New()

	// Use common middlerwares
	router.Use(log.LoggerMiddlerware(nil))
	router.Use(gin.Recovery())

	// Register handlers
	register(router)

	return router
}

func register(router *gin.Engine) {
	root(router.Group("/"))
	goffer.Register(router.Group("/goffer"))
}

// root
func root(router *gin.RouterGroup) {
	router.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
}
