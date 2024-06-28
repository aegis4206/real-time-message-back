package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
	Username string
	Uuid     string
	Obs      bool
	Mutex    sync.Mutex
}

func (ws *WebSocketConnection) WriteJSON(v interface{}) error {
	ws.Mutex.Lock()
	defer ws.Mutex.Unlock()
	return ws.Conn.WriteJSON(v)
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
	File    string
	Id      string
	To      string
	ToUser  string
	RTCType string
	List    []*WebSocketConnection
}

// type ConnectedResponse struct {
// 	From string
// 	Type string
// 	List []*WebSocketConnection
// 	Id   string
// }

var (
	connections = make([]*WebSocketConnection, 0)
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "https://192.168.6.87:8006"
		},
	}
	base64DataURIRegex = regexp.MustCompile(`^data:([a-zA-Z0-9]+/[a-zA-Z0-9-.+]+);base64,(.*)$`)
)

func UserConnectedOrDis(currentConn *WebSocketConnection, connections []*WebSocketConnection, status bool) {
	state := "disconnected"
	if status {
		state = "connected"
	}

	socketResponse := SocketResponse{
		From: currentConn.Username,
		Type: state,
		Id:   currentConn.Uuid,
		List: connections,
	}

	for _, v := range connections {
		if err := v.WriteJSON(socketResponse); err != nil {
			return
		}
	}
	socketResponse.Type = "id"
	currentConn.WriteJSON(socketResponse)

	fmt.Println(currentConn.Username)
	fmt.Println("Current Connection: ", len(connections))
}

func InitWs(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	conn, _ := upgrader.Upgrade(w, r, nil)
	uuid, _ := uuid.NewV4()
	id := uuid.String()
	currentConn := &WebSocketConnection{Conn: conn, Username: username, Uuid: id, Obs: false}
	connections = append(connections, currentConn)

	go UserConnectedOrDis(currentConn, connections, true)

	go handleMessages(conn, currentConn)
	// for {
	// 	// msgType, msg, err := conn.ReadMessage()
	// 	var messageRep SocketResponse
	// 	err := conn.ReadJSON(&messageRep)
	// 	if err != nil {
	// 		if err != nil {
	// 			var tempConnections = make([]*WebSocketConnection, 0)

	// 			for _, v := range connections {
	// 				if currentConn.Conn != v.Conn {
	// 					tempConnections = append(tempConnections, v)
	// 				}
	// 			}
	// 			connections = tempConnections
	// 			UserConnectedOrDis(currentConn, connections, false)
	// 			return
	// 		}

	// 		return
	// 	}

	// 	// message := string(msg)
	// 	// var messageRep SocketResponse
	// 	// json.Unmarshal(msg, &messageRep)
	// 	// var resp SocketResponse
	// 	matches := base64DataURIRegex.FindStringSubmatch(messageRep.Message)
	// 	if messageRep.Type == "file" {
	// 		mimeType := matches[1]
	// 		// base64Data := matches[2]
	// 		// data, err := base64.StdEncoding.DecodeString(base64Data)
	// 		fmt.Printf("file type: %s\n", mimeType)
	// 		messageRep = SocketResponse{
	// 			From:    messageRep.From,
	// 			Type:    mimeType,
	// 			Message: messageRep.Message,
	// 			File:    messageRep.File,
	// 			Id:      messageRep.Id,
	// 			To:      messageRep.To,
	// 			ToUser:  messageRep.ToUser,
	// 			RTCType: messageRep.RTCType,
	// 		}
	// 	} else {
	// 		// fmt.Printf("%s %s: %s\n", conn.RemoteAddr(), username, messageRep.Message)
	// 		messageRep = SocketResponse{
	// 			From:    messageRep.From,
	// 			Type:    messageRep.Type,
	// 			Message: messageRep.Message,
	// 			File:    "",
	// 			Id:      messageRep.Id,
	// 			To:      messageRep.To,
	// 			ToUser:  messageRep.ToUser,
	// 			RTCType: messageRep.RTCType,
	// 		}

	// 	}
	// 	// byteResp, _ := json.Marshal(messageRep)

	// 	// 更新主持人
	// 	if messageRep.Type == "obs" {
	// 		for _, v := range connections {
	// 			if v.Uuid == messageRep.Id {
	// 				if messageRep.Message == "start" {
	// 					v.Obs = true
	// 				} else {
	// 					v.Obs = false
	// 				}
	// 			}
	// 		}
	// 		messageRep.List = connections
	// 	}

	// 	for _, v := range connections {

	// 		// if err = v.Conn.WriteJSON(messageRep); err != nil {
	// 		// 	return
	// 		// }

	// 		if messageRep.To != "" {
	// 			if currentConn.Uuid == v.Uuid || messageRep.To == v.Uuid {
	// 				if err = v.WriteJSON(messageRep); err != nil {
	// 					return
	// 				}
	// 			}

	// 		} else {
	// 			if err = v.WriteJSON(messageRep); err != nil {
	// 				return
	// 			}
	// 		}

	// 	}
	// }
}

func handleMessages(conn *websocket.Conn, currentConn *WebSocketConnection) {
	for {
		var messageRep SocketResponse
		err := conn.ReadJSON(&messageRep)
		if err != nil {
			if err != nil {
				var tempConnections = make([]*WebSocketConnection, 0)

				for _, v := range connections {
					if currentConn.Conn != v.Conn {
						tempConnections = append(tempConnections, v)
					}
				}
				connections = tempConnections
				go UserConnectedOrDis(currentConn, connections, false)
				return
			}

			return
		}

		matches := base64DataURIRegex.FindStringSubmatch(messageRep.Message)
		if messageRep.Type == "file" {
			mimeType := matches[1]
			fmt.Printf("file type: %s\n", mimeType)
			messageRep = SocketResponse{
				From:    messageRep.From,
				Type:    mimeType,
				Message: messageRep.Message,
				File:    messageRep.File,
				Id:      messageRep.Id,
				To:      messageRep.To,
				ToUser:  messageRep.ToUser,
				RTCType: messageRep.RTCType,
			}
		} else {
			// fmt.Printf("%s %s: %s\n", conn.RemoteAddr(), username, messageRep.Message)
			messageRep = SocketResponse{
				From:    messageRep.From,
				Type:    messageRep.Type,
				Message: messageRep.Message,
				File:    "",
				Id:      messageRep.Id,
				To:      messageRep.To,
				ToUser:  messageRep.ToUser,
				RTCType: messageRep.RTCType,
			}

		}

		// 更新主持人
		if messageRep.Type == "obs" {
			for _, v := range connections {
				if v.Uuid == messageRep.Id {
					if messageRep.Message == "start" {
						v.Obs = true
					} else {
						v.Obs = false
					}
				}
			}
			messageRep.List = connections
		}

		for _, v := range connections {
			if messageRep.To != "" {
				if currentConn.Uuid == v.Uuid || messageRep.To == v.Uuid {
					if err = v.WriteJSON(messageRep); err != nil {
						return
					}
				}
			} else {
				if err = v.WriteJSON(messageRep); err != nil {
					return
				}
			}
		}
	}
}
