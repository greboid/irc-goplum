package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/greboid/irc/v2/logger"
	"github.com/greboid/irc/v2/rpc"
	"github.com/kouhin/envflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"net/http"
	"time"
)

var (
	RPCHost  = flag.String("rpc-host", "localhost", "gRPC server to connect to")
	RPCPort  = flag.Int("rpc-port", 8001, "gRPC server port")
	RPCToken = flag.String("rpc-token", "", "gRPC authentication token")
	Channel  = flag.String("channel", "", "Channel to send messages to")
	Secret   = flag.String("secret", "", "Secret for masking url")
	Debug    = flag.Bool("debug", false, "Show debugging info")
)

type goplum struct {
	client rpc.IRCPluginClient
	log    *zap.SugaredLogger
}

func main() {
	log := logger.CreateLogger(*Debug)
	if err := envflag.Parse(); err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
	}
	github := goplum{
		log: log,
	}
	log.Infof("Creating goplum RPC Client")
	client, err := github.doRPC()
	if err != nil {
		log.Fatalf("Unable to create RPC Client: %s", err.Error())
	}
	github.client = client
	log.Infof("Starting goplum web server")
	err = github.doWeb()
	if err != nil {
		log.Panicf("Error handling web: %s", err.Error())
	}
	log.Infof("exiting")
}

func (g *goplum) doRPC() (rpc.IRCPluginClient, error) {
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *RPCHost, *RPCPort), grpc.WithTransportCredentials(creds))
	client := rpc.NewIRCPluginClient(conn)
	_, err = client.Ping(rpc.CtxWithToken(context.Background(), "bearer", *RPCToken), &rpc.Empty{})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (g *goplum) doWeb() error {
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *RPCHost, *RPCPort), grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	client := rpc.NewHTTPPluginClient(conn)
	stream, err := client.GetRequest(rpc.CtxWithTokenAndPath(context.Background(), "bearer", *RPCToken, "goplum"))
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}
		response := g.handleGoPlum(request)
		err = stream.Send(response)
		if err != nil {
			return err
		}
	}
}

func (g *goplum) handleGoPlum(request *rpc.HttpRequest) *rpc.HttpResponse {
	g.log.Infof("Received webhook: %s vs %s", request.Path, fmt.Sprintf("goplum/%s", *Secret))
	if request.Path != fmt.Sprintf("/goplum/%s", *Secret) {
		return &rpc.HttpResponse{
			Status: http.StatusNotFound,
			Body:   []byte("Not found."),
		}
	}
	go func() {
		data := GoPlumHook{}
		err := json.Unmarshal(request.Body, &data)
		if err != nil {
			g.log.Errorf("Unable to handle webhook: %s", err.Error())
		}

		g.sendMessage([]string{fmt.Sprintf("Monitoring: %s", data.Text)})
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

func (g *goplum) sendMessage(messages []string) []error {
	errors := make([]error, 0)
	for index := range messages {
		_, err := g.client.SendChannelMessage(rpc.CtxWithToken(context.Background(), "bearer", *RPCToken), &rpc.ChannelMessage{
			Channel: *Channel,
			Message: messages[index],
		})
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
