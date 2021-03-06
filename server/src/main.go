package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"ren0503/talkive/src/controllers"
	"ren0503/talkive/src/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var sockets = make(map[string]map[string]*models.Connection)

func wshandler(w http.ResponseWriter, r *http.Request, socket string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error handling websocket connection.")
		return
	}

	defer conn.Close()

	if sockets[socket] == nil {
		sockets[socket] = make(map[string]*models.Connection)
	}

	clients := sockets[socket]

	var message models.Message
	for {
		err = conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if clients[message.UserID] == nil {
			connection := new(models.Connection)
			connection.Socket = conn
			clients[message.UserID] = connection
		}

		switch message.Type {
		case "connect":
			message.Type = "session_joined"
			err := conn.WriteJSON(message)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				delete(clients, message.UserID)
			}
			break
		case "disconnect":
			for user, client := range clients {
				err := client.Send(message)
				if err != nil {
					client.Socket.Close()
					delete(clients, user)
				}
			}
			delete(clients, message.UserID)
			break
		default:
			for user, client := range clients {
				err := client.Send(message)
				if err != nil {
					delete(clients, user)
				}
			}
		}
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{getenv("HOST_URL", "localhost")}

	router.Use(cors.Default())

	credential := options.Credential{
		Username: "user0503",
		Password: "116105101110",
	}

	clientOptions := options.Client().ApplyURI("URL").SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("MongoDB connection ok...")

	// middleware - intercept requests to use our db controller
	router.Use(func(context *gin.Context) {
		context.Set("db", client)
		context.Next()
	})

	// REST API
	router.POST("/session", controllers.CreateSession)
	router.GET("/connect", controllers.GetSession)
	router.POST("/connect/:url", controllers.ConnectSession)

	// Websocket connection
	router.GET("/ws/:socket", func(c *gin.Context) {
		socket := c.Param("socket")
		wshandler(c.Writer, c.Request, socket)
	})

	router.Run("0.0.0.0:" + getenv("PORT", "5000"))
}
