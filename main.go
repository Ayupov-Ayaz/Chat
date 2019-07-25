// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"chat/configs"
	"chat/core"
	"flag"
	"go.uber.org/zap"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// front
	http.ServeFile(w, r, "public/index.html")
}

func main() {
	flag.Parse()
	server := core.NewServer()
	go server.Run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(server, w, r)
	})
	if err := http.ListenAndServe(*addr, nil); err != nil {
		zap.L().Named("main").Error("lister and server failed", zap.Error(err))
	}
}

func serveWs(server *core.Server, w http.ResponseWriter, r *http.Request) {
	l := zap.L().Named("serverWsHandler")
	conn, err := configs.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Error("ws upgrade failed", zap.Error(err))
		return
	}
	client := core.MakeClient(server, conn)
	client.Subscribe()
	go client.WritePump()
	go client.ReadPump()
}