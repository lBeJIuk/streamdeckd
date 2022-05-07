package main

import (
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/unix-streamdeck/api"
	"log"
	"net/http"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}

func InitWS() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("Error on ws upgrade: %v", err)
		}
		go func() {
			defer conn.Close()

			for {
				rawMsg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					log.Printf("Error on ReadClientData: %v", err)
				}
				msg := new(Message)
				err = json.Unmarshal(rawMsg, msg)
				if err != nil {
					log.Printf("Error on Unmarshal: %v", err)
					return
				}
				switch msg.Type {
				case "getConfig":
					configString, _ := json.Marshal(config)
					resp := Message{
						Type: "getConfig",
						Data: string(configString),
					}
					respString, _ := json.Marshal(resp)
					err = wsutil.WriteServerMessage(conn, op, respString)
					if err != nil {
						log.Printf("Error on WriteServerMessage: %v", err)
						return
					}
					break
				case "setConfig":
					newConfig := new(api.Config)
					err = json.Unmarshal([]byte(msg.Data), newConfig)
					if err != nil {
						log.Printf("Error on Unmarshal: %v", err)
						return
					}
					config = newConfig
					_ = SaveConfig()
					_ = ReloadConfig()
					break
				}
			}
		}()
	}))
}
