package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)
type Operation struct {
	Type     string `json:"type"`
	Position int    `json:"position"`
	OldText  string `json:"oldText"`
	NewText  string `json:"newText"`
	ClientId string `json:"clientId"`
}

type clientInfo struct {
	conn *websocket.Conn
	clientId string
}

var clients = make(map[*websocket.Conn]string)
var addClient = make(chan clientInfo)
var removeClient = make(chan *websocket.Conn)
var broadcasts = make(chan Operation)
var documentState string;


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true //allows any client to connect
	},
}

func broadcaster(){
	
	for  {
		select{
		case client := <- addClient:
			clients[client.conn] = client.clientId
		case conn:= <- removeClient:
			delete(clients, conn)
		case op:= <- broadcasts:
			opJson, err := json.Marshal(op)
			if err != nil {
				log.Printf("Failed to marshal operation: %v", err)
				continue
			}
			for client, clientId := range clients{
				if clientId != op.ClientId{
					err := client.WriteMessage(websocket.TextMessage,opJson )
					if err != nil{
						log.Printf("error: %v", err )
						client.Close()
						delete(clients, client)
					}

				}

			}
	}

	}
}

func handler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
	defer conn.Close()
	clientId := uuid.New().String()

	conn.WriteJSON(map[string]interface{}{
		"type": "assign_client_id",
		"clientId": clientId,
	})

	conn.WriteJSON(map[string]interface{}{
        "type": "current_state",
        "content": documentState,
    })
	addClient <- clientInfo{conn: conn, clientId: clientId}

	for  {
		

	mt, msg, err := conn.ReadMessage()
	if err != nil{
		log.Printf("message failed to read: %v ", err)
		delete(clients, conn)
		break
	}
	var op Operation
	err =json.Unmarshal(msg, &op)
	if err != nil {
		log.Printf("Failed to unmarshal operation: %v", err)
		return
	}
	documentState = applyChange(documentState, op)


	log.Printf("Message of type %v received", mt)
	broadcasts <- op

	
}

}

func applyChange(content string, change Operation) string {
    return content[:change.Position] + change.NewText + content[change.Position+len(change.OldText):]
}





func main(){
	http.HandleFunc("/streamer" , handler)
	go broadcaster()
	log.Println("Server started and listening on :8080...") 
	log.Fatal(http.ListenAndServe(":8080", nil))
}