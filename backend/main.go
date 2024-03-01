package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	sloggin "github.com/samber/slog-gin"
)

// Response represents the structure of the API response.
type Response struct {
	Success bool   `json:"success"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

var mqttClient mqtt.Client

func main() {
	Init()
	setupLogger()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(sloggin.New(slog.Default()), gin.Recovery())
	router.POST("/api/open", handleOpen)
	router.POST("/api/log", handleLog)

	connectToMQTTBroker()

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %+v", err)
	}
}

func handleOpen(c *gin.Context) {
	type Request struct {
		Action   string `json:"action"`
		ToServer struct {
			ShortResponse bool `json:"shortResponse"`
		} `json:"toServer"`
		Passthrough json.RawMessage `json:"passthrough"`
	}
	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Info("c.ShouldBindJSON() failed", slogError(err))
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: "Invalid request"})
		return
	}

	if req.Action != "open" {
		slog.Info("Invalid action", slog.String("action", req.Action))
		c.JSON(http.StatusNotImplemented, Response{Success: false, Message: "Action not implemented"})
		return
	}

	if err := publishMQTTMessage("siyangz/home/gate",
		map[string]any{"command": "open", "passthrough": req.Passthrough}); err != nil {
		slog.Info("Failed to publish MQTT message", slogError(err))
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: "Failed to open door"})
		return
	}

	slog.Info("Door unlocked", slog.String("action", req.Action))
	c.JSON(http.StatusOK, Response{Success: true, Message: "Door unlocked"})
}

func handleLog(c *gin.Context) {
	type Request struct {
		Event string `json:"event"` // like gate_open
	}
	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Info("c.ShouldBindJSON() failed", slogError(err))
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: "Invalid request"})
		return
	}

	slog.Info("Event logged", slog.String("event", req.Event))

	c.JSON(http.StatusOK, Response{Success: true, Message: "Log recorded"})
}

func Init() {
	gin.SetMode(gin.ReleaseMode)
}

func setupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)
}

func connectToMQTTBroker() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://broker.hivemq.com:1883")
	opts.SetClientID("iot-backend")
	opts.SetWriteTimeout(5 * time.Second)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(true)

	mqttClient = mqtt.NewClient(opts)
	token := mqttClient.Connect()
	token.Wait()

	if token.Error() != nil {
		slog.Error("Failed to connect to MQTT broker", slogError(token.Error()))
		os.Exit(1)
		return
	}

	slog.Info("Connected to MQTT broker")
}

func publishMQTTMessage(topic string, payload any) (err error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal MQTT payload", slogError(err))
		return errors.Wrapf(err, "failed to marshal MQTT payload: %+v", payload)
	}

	token := mqttClient.Publish(topic, 1, false, payloadBytes)
	token.Wait()

	if token.Error() != nil {
		slog.Error("Failed to publish MQTT message", slogError(token.Error()))
		return errors.Wrapf(token.Error(), "failed to publish MQTT message, topic: %s, payload: %s", topic, payloadBytes)
	}

	return nil
}

func slogError(e error) slog.Attr {
	return slog.Any("err", e)
}
