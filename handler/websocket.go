package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
	Username string
	Uuid     string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
	File    string
	Id      string
	To      string
	ToUser  string
}
type connectedResponse struct {
	From string
	Type string
	List []*WebSocketConnection
	Id   string
}

var (
	connections = make([]*WebSocketConnection, 0)
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "http://192.168.6.87:8006"
		},
	}
	base64DataURIRegex = regexp.MustCompile(`^data:([a-zA-Z0-9]+/[a-zA-Z0-9-.+]+);base64,(.*)$`)
)

func UserConnectedOrDis(currentConn WebSocketConnection, connections []*WebSocketConnection, status bool) {
	state := "disconnected"
	if status {
		state = "connected"
	}

	connectedBypeResp, _ := json.Marshal(connectedResponse{
		From: currentConn.Username,
		Type: state,
		Id:   currentConn.Uuid,
		List: connections,
	})

	for _, v := range connections {
		// if currentConn.Conn != v.Conn {
		// 	if err := v.Conn.WriteMessage(1, connectedBypeResp); err != nil {
		// 		return
		// 	}
		// }
		if err := v.Conn.WriteMessage(1, connectedBypeResp); err != nil {
			return
		}
	}

	fmt.Println(currentConn.Username)
	fmt.Println("Current Connection: ", len(connections))
}

func InitWs(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	conn, _ := upgrader.Upgrade(w, r, nil)
	uuid, _ := uuid.NewV4()
	id := uuid.String()
	currentConn := WebSocketConnection{Conn: conn, Username: username, Uuid: id}
	connections = append(connections, &currentConn)

	UserConnectedOrDis(currentConn, connections, true)

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if err != nil {
				var tempConnections = make([]*WebSocketConnection, 0)

				for _, v := range connections {
					if currentConn.Conn != v.Conn {
						tempConnections = append(tempConnections, v)
					}
				}
				connections = tempConnections
				UserConnectedOrDis(currentConn, connections, false)
				return
			}

			return
		}

		// message := string(msg)
		var messageRep SocketResponse
		json.Unmarshal(msg, &messageRep)
		// var resp SocketResponse
		matches := base64DataURIRegex.FindStringSubmatch(messageRep.Message)
		if messageRep.File != "" {
			mimeType := matches[1]
			// base64Data := matches[2]
			// data, err := base64.StdEncoding.DecodeString(base64Data)
			if err == nil {
				fmt.Printf("file type: %s\n", mimeType)
				messageRep = SocketResponse{
					From:    currentConn.Username,
					Type:    mimeType,
					Message: string(messageRep.Message),
					File:    messageRep.File,
					Id:      messageRep.Id,
					To:      messageRep.To,
					ToUser:  messageRep.ToUser,
				}
			}
		} else {
			// fmt.Printf("%s %s: %s\n", conn.RemoteAddr(), username, messageRep.Message)
			messageRep = SocketResponse{
				From:    currentConn.Username,
				Type:    "message",
				Message: string(messageRep.Message),
				File:    "",
				Id:      messageRep.Id,
				To:      messageRep.To,
				ToUser:  messageRep.ToUser,
			}

		}
		byteResp, _ := json.Marshal(messageRep)

		for _, v := range connections {

			if messageRep.To != "" {

				if currentConn.Uuid == v.Uuid || messageRep.To == v.Uuid {
					if err = v.Conn.WriteMessage(msgType, byteResp); err != nil {
						return
					}
				}

			} else {
				if err = v.Conn.WriteMessage(msgType, byteResp); err != nil {
					return
				}
			}

		}
	}
}
