package main

import (
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net/http"
)

func InitWS() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("Error on ws upgrade: %v", err)
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					log.Printf("Error on ReadClientData: %v", err)
				}
				log.Printf("msg: %v", msg)

				configString, err := json.Marshal(config)
				if err != nil {
					log.Printf("Error on config marshaling: %v", err)
				}

				err = wsutil.WriteServerMessage(conn, op, configString)
				if err != nil {
					log.Printf("Error on WriteServerMessage: %v", err)
				}
			}
		}()
	}))
}
