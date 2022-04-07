package socketio

import (
	"memphis-server/logger"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

type myResponse struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func InitializeSocketio(router *gin.Engine) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		logger.Info("A new socket created " + s.ID())
		return nil
	})

	server.OnEvent("/endpoint", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		a := myResponse{Field1: "Idan Asulin", Field2: 4}
		s.Emit("msg_1", a)
		return "recv " + msg
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Error("An error occured " + e.Error())
	})

	message := make(chan string)
	go func() {
		if err := server.Serve(); err != nil {
			logger.Error("socketio listen error " + err.Error())
		}
		message <- "ping"
	}()

	// msg := <- message
	// fmt.Println(msg)

	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

	return server
}
