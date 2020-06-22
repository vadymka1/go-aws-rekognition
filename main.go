package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
	"upload_images/services"
)

type Server struct {
	http.Server
	shutdownReq chan bool
	reqCount uint32
}

var port = ":9090"
var word string

func main() {
	server := StartServer()

	done := make(chan bool)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Listen and serve: %v", err)
		}
		log.Println("Server starting")
		done <- true
	}()

	server.WaitShutDown()

	<-done
	log.Printf("DONE!")
}

func StartServer() *Server {
	srv := &Server{
		Server:      http.Server{
			Addr         : port,
			ReadTimeout  : 10*time.Second,
			WriteTimeout : 10*time.Second,
		},
		shutdownReq: make(chan bool),
		reqCount:    0,
	}

	router := mux.NewRouter()
	router.HandleFunc("/shutdown", srv.ShutdownHandler)
	router.HandleFunc("/showform", services.GetUploadForm).Methods("GET")
	router.HandleFunc("/upload", services.UploadHandler).Methods("POST")

	srv.Handler = router

	return srv
}

func (s *Server) WaitShutDown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v", sig)
	case sig := <-s.shutdownReq:
		log.Printf("Shutdown request (/shutdown %v)", sig)
	}

	log.Println("Stopping http server..")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v", err)
	}

}

func (s *Server) ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Shutdown server"))

	if !atomic.CompareAndSwapUint32(&s.reqCount, 0, 1) {
		log.Printf("Shutdown through API call in progress...")
		return
	}

	go func() {
		s.shutdownReq <- true
	}()
}