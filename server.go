package main

import (
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

var psClient *pubsub.Client
var psStartComputationTopic *pubsub.Topic
var dsClient *datastore.Client

func init() {
	var err error
	ctx := context.Background()

	psClient, err = pubsub.NewClient(ctx, "paul-test-321412")
	if err != nil {
		log.Fatal(err)
	}
	psStartComputationTopic = psClient.Topic("compute")
	dsClient, err = datastore.NewClient(ctx, "paul-test-321412")
	if err != nil {
		panic(err)
	}
}

func main() {
	// init server
	server := echo.New()
	port := ":8000"

	server.Validator = &StructValidator{validator: validator.New()}

	// init handlers
	server.POST("/", addWebhook)
	server.POST("/:id/computation", computeWebhook)
	server.PUT("/:id", editWebhook)
	server.DELETE("/:id", deleteWebhook)
	server.GET("/:id", getWebhook)
	server.GET("/", getAllWebhook)
	server.GET("/:webhookId/computation/:computationId", getComputationResult)
	server.GET("/close", func(context echo.Context) error {
		if err := server.Close(); err != nil {
			return err
		}
		return echo.NewHTTPError(http.StatusOK, "closing the server")
	})

	// start server
	if err := server.Start(port); err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// end server
	defer func() {
		_ = dsClient.Close()
		psStartComputationTopic.Stop()
		_ = psClient.Close()
	}()
}
