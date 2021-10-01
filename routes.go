package main

import (
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func addWebhook(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	newWebhook := new(Webhook)

	if err := getRequestBody(c, newWebhook); err != nil {
		return err
	}
	if len(newWebhook.Fields) <= 0 {
		return c.String(http.StatusBadRequest, "fields cannot be empty")
	}
	key, err := dsClient.Put(
		ctx,
		datastore.IncompleteKey("Webhook", nil),
		newWebhook)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	newWebhook.ID = key.ID
	return c.JSON(http.StatusOK, newWebhook)
}

func computeWebhook(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	compute := new(Computation)
	valueMap := new(ComputationBody)
	webhook := new(Webhook)
	buffer := 0

	if _, err := getEntityById(ctx, c.Param("id"), "Webhook", webhook); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := getRequestBody(c, valueMap); err != nil {
		return err
	}
	if len(valueMap.ValueMap) == 0 {
		return c.String(http.StatusBadRequest, "values cannot be empty")
	}
	compute.WebhookID = webhook.ID
	compute.Computed = false
	compute.Result = nil
	compute.Values = make([]ComputationValues, len(valueMap.ValueMap))
	for i, j := range valueMap.ValueMap {
		compute.Values[buffer] = ComputationValues{
			Key:   i,
			Value: j,
		}
		buffer++
	}
	key, err := dsClient.Put(
		ctx,
		datastore.IncompleteKey("Computation", nil),
		compute)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	compute.ID = key.ID
	pubsubMessageData, _ := json.Marshal(PubSubPayload{
		ComputationId: compute.ID,
		Op:            webhook.Op,
		Fields:        webhook.Fields,
		Values:        valueMap.ValueMap,
	})
	_ = psStartComputationTopic.Publish(ctx, &pubsub.Message{
		Data: pubsubMessageData,
	})
	return c.JSON(http.StatusOK, compute)
}

func editWebhook(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var err error
	var key *datastore.Key
	webhook := new(Webhook)

	if key, err = getEntityById(ctx, c.Param("id"), "Webhook", webhook); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = getRequestBody(c, webhook); err != nil {
		return err
	}
	if _, err = dsClient.Put(ctx, key, webhook); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, webhook)
}

func deleteWebhook(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var err error
	var key *datastore.Key
	var keys []*datastore.Key

	if key, err = getEntityById(ctx, c.Param("id"), "Webhook", new(Webhook)); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if keys, err = getAllWebhookComputation(ctx, key.ID); err != nil {
		return err
	}
	keys = append(keys, key)
	if err = dsClient.DeleteMulti(ctx, keys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Webhook successfully delete")
}

func getWebhook(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	webhook := new(Webhook)

	if _, err := getEntityById(ctx, c.Param("id"), "Webhook", webhook); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return echo.NewHTTPError(http.StatusBadRequest, webhook)
}

func getAllWebhook(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var webhooks []*Webhook

	_, err := dsClient.GetAll(
		ctx,
		datastore.NewQuery("Webhook"),
		&webhooks)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, webhooks)
}

func getComputationResult(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	computation := new(Computation)

	key, err := getEntityById(ctx, c.Param("computationId"), "Computation", computation)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	computation.ID = key.ID
	return c.JSON(http.StatusOK, computation)
}
