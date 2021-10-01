package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func getEntityById(ctx context.Context, stringId string, kind string, i interface{}) (*datastore.Key, error) {
	var err error
	var id int64 = 0

	if id, err = strconv.ParseInt(stringId, 10, 64); err != nil {
		return nil, err
	}
	key := &datastore.Key{
		Kind:      kind,
		ID:        id,
		Name:      "",
		Parent:    nil,
		Namespace: "",
	}
	if err = dsClient.Get(ctx, key, i); err != nil {
		return key, err
	}
	return key, nil
}

func getAllWebhookComputation(ctx context.Context, webhookID int64) ([]*datastore.Key, error) {
	query := datastore.
		NewQuery("Computation").
		Filter("webhookId =", webhookID).
		KeysOnly()

	if keys, err := dsClient.GetAll(ctx, query, nil); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else {
		return keys, nil
	}
}

func getRequestBody(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
