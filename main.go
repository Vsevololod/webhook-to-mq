package main

import (
	"log/slog"
	"net/http"
	"os"
	"webhook-to-mq/config"
	"webhook-to-mq/lib/sl"

	"github.com/go-chi/chi/v5"
	"github.com/rabbitmq/amqp091-go"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	rabbitURL := os.Getenv(cfg.AmqpConf.GetAmqpUri())
	log.Info("Try to connect to uri=" + rabbitURL)
	rabbitConn, err := amqp091.Dial(rabbitURL)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ:", sl.Err(err))
	}
	defer rabbitConn.Close()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Error("Failed to open a channel:", sl.Err(err))
	}
	defer ch.Close()

	exchangeName := "webhooks"
	err = ch.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("Failed to declare exchange:", sl.Err(err))
	}

	r := chi.NewRouter()
	r.Post("/webhook/{senderName}", func(w http.ResponseWriter, r *http.Request) {
		senderName := chi.URLParam(r, "senderName")

		body := make([]byte, r.ContentLength)
		_, err := r.Body.Read(body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		err = ch.Publish(
			exchangeName,
			senderName,
			false,
			false,
			amqp091.Publishing{
				ContentType: r.Header.Get("Content-Type"),
				Body:        body,
			},
		)
		if err != nil {
			http.Error(w, "Failed to publish message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Info("Starting webhook service on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Error("Server failed:", sl.Err(err))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
