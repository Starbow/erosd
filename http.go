package main

import (
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"

	"errors"
	"os"

	"html/template"
)

var (
	tlsCertPath string
	tlsKeyPath  string

	webRoot = "web"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebsocketConnection(conn *websocket.Conn) {
	client := NewWebsocketClientConnection(*conn)
	client.read()
}

func httpServeHomeFunc(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, webRoot+"/dist/index.html")
}

func httpServeCharVerification(w http.ResponseWriter, r *http.Request) {
	type Page struct {
		Char  Sc2Char
		Error string
	}

	var (
		p    Page
		err  error
		char Sc2Char
	)

	state := r.URL.Query()["state"][0]

	// Get the OAuth for the state requested
	if oar, ok := activeOAuths[state]; ok {
		var proto_char *BattleNetCharacter

		char, proto_char, err = oar.getCharInfo(r.URL.Query()["code"][0])

		if err != nil {
			oar.conn.logger.Println(err)
		}

		payload := proto_char.CharacterMessage()
		data, _ := Marshal(payload)
		oar.conn.SendResponseMessage("BNN", -1, data)
	} else {
		err = errors.New("This is not a valid request, please try again.")
	}

	t, templ_err := template.ParseFiles(webRoot + "dist/verify_char.html")

	if templ_err != nil {
		log.Println(templ_err)
	}
	if err != nil {
		p.Error = err.Error()
	}

	p.Char = char

	t.Execute(w, p)
}

func httpServeWsFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade WS from", r.RemoteAddr, err)
		return
	}

	log.Println("Accepted WS from", ws.RemoteAddr())
	go handleWebsocketConnection(ws)
}

func setupRouter(r *mux.Router) {
	r.HandleFunc("/ws", httpServeWsFunc).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(webRoot+"/dist")))).Methods("GET")
	r.HandleFunc("/login/battlenet", httpServeCharVerification).Methods("GET")
	r.HandleFunc("/{path:.*}", httpServeHomeFunc).Methods("GET")
}

func listenAndServeHTTP(address string) {
	r := mux.NewRouter()
	setupRouter(r)

	log.Println("Listening HTTP on", address)
	go http.ListenAndServe(address, r)
	// log.Fatalln()
}

func listenAndServeHTTPS(address string) error {
	if _, err := os.Stat(tlsKeyPath); err == nil {
		r := mux.NewRouter()
		setupRouter(r)

		srv := &http.Server{Addr: address, Handler: r}
		addr := srv.Addr

		config := &tls.Config{}
		if srv.TLSConfig != nil {
			*config = *srv.TLSConfig
		}
		if config.NextProtos == nil {
			config.NextProtos = []string{"http/1.1"}
		}

		var err error
		config.Certificates = make([]tls.Certificate, 1)

		config.Certificates[0], err = tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)

		if err != nil {
			log.Fatalln(err)
		}

		ln, err := net.Listen("tcp", addr)

		if err != nil {
			log.Fatalln(err)
		}

		tlsListener := tls.NewListener(ln, config)

		go srv.Serve(tlsListener)

		log.Println("Listening HTTPS on", address)
	} else {
		log.Fatalln("Could not use https certs: ", err)
	}
	return nil
}
