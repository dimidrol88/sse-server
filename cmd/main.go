package main

import (
	"encoding/json"
	"fmt"
	"github.com/dimidrol88/sse-server/internal"
	"github.com/google/uuid"
	"log"
	"net/http"
)

const event = "example"

func subscribe(broker *internal.Broker) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		flusher, ok := res.(http.Flusher)
		if !ok {
			http.Error(res, "Streaming не поддерживается!", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/event-stream")
		res.Header().Set("Cache-Control", "no-cache")
		res.Header().Set("Connection", "keep-alive")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		//res.Header().Set("Access-Control-Allow-Origin", "http://localhost:63342")
		//res.Header().Set("Access-Control-Allow-Credentials", "true")

		connect := &internal.Connect{
			Id:        uuid.NewString(),
			Event:     event,
			MessageCh: make(chan *internal.Message, 10),
			PingCh:    make(chan string, 1),
			DoneCh:    make(chan struct{}),
		}

		broker.Add(connect)
		defer broker.Remove(connect.Id)

		broker.Ping(connect)

		for {
			select {
			case msg := <-connect.PingCh:
				_, _ = fmt.Fprintf(
					res,
					"retry: %d\nevent: %s\ndata: %s\n\n",
					5000,
					"ping",
					msg,
				)
				flusher.Flush()
			case msg := <-connect.MessageCh:
				jsonData, err := json.Marshal(msg)
				if err != nil {
					log.Println("Ошибка при преобразовании в JSON:", err)
					continue
				}

				_, _ = fmt.Fprintf(
					res,
					"event: %s\ndata: %s\n\n",
					connect.Event,
					jsonData,
				)
				flusher.Flush()
			case <-req.Context().Done():
				_, _ = fmt.Fprintf(res, "data: %s\n\n", "Пользователь закрыл страницу!")
				return
			case <-connect.DoneCh:
				_, _ = fmt.Fprintf(res, "data: %s\n\n", "Пользователя отключил сервер!")
				return
			}
		}
	}
}

func publish(broker *internal.Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("message")
		if msg == "" {
			http.Error(w, "Параметр 'message' обязателен!", http.StatusBadRequest)
			return
		}

		broker.Notify(&internal.Message{Message: msg})

		_, _ = w.Write([]byte("200"))
	}
}

func main() {
	broker := internal.NewBroker()

	http.HandleFunc("/subscribe", subscribe(broker))
	http.HandleFunc("/publish", publish(broker))

	log.Println("SSE server запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
