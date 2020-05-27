package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	"golang.org/x/net/websocket"
)

var Users map[int64]*User

type Msg struct {
	ID       int64
	FromID   int64
	Time     int64
	Content  string
	Receiver []int64
}

type User struct {
	ID      int64
	Name    string
	Channel chan Msg
}

func init() {
	Users = make(map[int64]*User)
	go http.ListenAndServe(":6060", nil)
}

func main() {
	http.HandleFunc("/web", Web)
	http.Handle("/websocket", websocket.Handler(Echo))
	log.Fatal(http.ListenAndServe(":81", nil))
}

func Echo(ws *websocket.Conn) {
	channel := make(chan Msg)

	id, _ := strconv.Atoi(ws.Request().URL.Query()["id"][0])
	if v, ok := Users[int64(id)]; ok {
		v.Channel = channel
	} else {
		return
	}
	go Work(ws, channel)
	for {
		var reply string
		if err := websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("err:", err)
			close(channel)
			break
		}
		msg := Msg{}
		json.Unmarshal([]byte(reply), &msg)
		msg.Time = time.Now().Unix()
		channel <- msg
	}
}

func Work(ws *websocket.Conn, channel chan Msg) {
	for {
		m, ok := <-channel
		if !ok {
			break
		}
		for _, v := range m.Receiver {
			if u, ok := Users[v]; ok && u.Channel != nil {
				if m.FromID == v {
					continue
				}
				Users[v].Channel <- Msg{
					ID:      0,
					Time:    time.Now().Unix(),
					Content: m.Content,
				}
			}
		}
		msg, _ := json.Marshal(m)
		if err := websocket.Message.Send(ws, string(msg)); err != nil {
			fmt.Println("err:", err)
			break
		}
	}

}

func Web(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method", r.Method)

	if r.Method == "GET" {
		name := r.URL.Query()["name"]
		id := r.URL.Query()["id"]
		atoi, _ := strconv.Atoi(id[0])

		Users[int64(atoi)] = &User{
			ID:   int64(atoi),
			Name: name[0],
		}

		t, _ := template.ParseFiles("websocket.html")

		t.Execute(w, map[string]interface{}{"Name": name, "ID": atoi})

	}
}
