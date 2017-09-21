package lib

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

var (
	R    *gin.Engine
	M    *melody.Melody
	MSGS chan []byte
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	R = gin.Default()
	M = melody.New()
}

func Run() {

	R.GET("/", func(c *gin.Context) {
		M.HandleRequest(c.Writer, c.Request)
	})

	M.HandleMessage(func(s *melody.Session, msg []byte) {
		M.Broadcast(msg)
	})
	R.Run(":2345")
}

func Write(str string) {
	MSGS <- []byte(str)
}
