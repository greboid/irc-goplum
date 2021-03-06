package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/greboid/irc-bot/v4/plugins"
	"github.com/greboid/irc-bot/v4/rpc"
	"github.com/greboid/irc/v4/logger"
	"github.com/kouhin/envflag"
	"go.uber.org/zap"
)

var (
	RPCHost  = flag.String("rpc-host", "localhost", "gRPC server to connect to")
	RPCPort  = flag.Int("rpc-port", 8001, "gRPC server port")
	RPCToken = flag.String("rpc-token", "", "gRPC authentication token")
	Channel  = flag.String("channel", "", "Channel to send messages to")
	Secret   = flag.String("secret", "", "Secret for masking url")
	Debug    = flag.Bool("debug", false, "Show debugging info")
	log      *zap.SugaredLogger
	helper   *plugins.PluginHelper
)

func main() {
	log = logger.CreateLogger(*Debug)
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
		return
	}
	log.Infof("Creating goplum RPC Client")
	helper, err = plugins.NewHelper(fmt.Sprintf("%s:%d", *RPCHost, uint16(*RPCPort)), *RPCToken)
	if err != nil {
		log.Fatalf("Unable to create plugin helper: %s", err.Error())
		return
	}
	err = helper.RegisterWebhook("goplum", handleGoPlum)
	if err != nil {
		log.Fatalf("Unable to register webhook: %s", err.Error())
		return
	}
	log.Infof("exiting")
}

func handleGoPlum(request *rpc.HttpRequest) *rpc.HttpResponse {
	if request.Path != fmt.Sprintf("/goplum/%s", *Secret) {
		return &rpc.HttpResponse{
			Status: http.StatusNotFound,
			Body:   []byte("Not found."),
		}
	}
	go func() {
		log.Infof("Received goplum notification")
		data := GoPlumHook{}
		err := json.Unmarshal(request.Body, &data)
		if err != nil {
			log.Errorf("Unable to handle webhook: %s", err.Error())
			return
		}
		if len(data.Text) == 0 {
			log.Debugf("Invalid webhook received")
			return
		}
		err = helper.SendChannelMessage(*Channel, fmt.Sprintf("Monitoring: %s", data.Text))
		if err != nil {
			log.Debugf("Error sending channel message: %s", err.Error())
		}
	}()
	return &rpc.HttpResponse{
		Body:   []byte("Delivered"),
		Status: http.StatusOK,
	}
}

type GoPlumHook struct {
	Text       string `json:"text"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	LastResult struct {
		State  string    `json:"state"`
		Time   time.Time `json:"time"`
		Detail string    `json:"detail"`
	} `json:"last_result"`
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
}
