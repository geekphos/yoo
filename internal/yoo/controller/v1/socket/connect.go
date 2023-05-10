package socket

import (
	"net/http"
	"strings"

	"phos.cc/yoo/internal/yoo/socket_client"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	"phos.cc/yoo/internal/pkg/log"
	"phos.cc/yoo/pkg/token"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ctrl *SocketController) Connect(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorw("upgrade", "err", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Errorw("read", "err", err)
			break
		}

		//  授权信息
		if strings.HasPrefix(string(message), "Bearer") {
			tokenString := strings.Replace(string(message), "Bearer ", "", 1)
			email, _, tokenType, _, err := token.Parse(tokenString, viper.GetString("jwt-secret"))
			if err != nil || tokenType != token.AccessToken {
				log.Errorw("token parse", "err", err)
				c.WriteMessage(mt, []byte("token parse error"))
				break
			}
			socket_client.AddConn(email, c)
		}

		log.Infow("recv", "mt", mt, "message", string(message))
	}
}
