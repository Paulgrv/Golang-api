package main

import (
	"cloud.google.com/go/datastore"
	"github.com/go-playground/validator"
)

type StructValidator struct {
	validator *validator.Validate
}

func (whv *StructValidator) Validate(i interface{}) error {
	if err := whv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

type Webhook struct {
	ID     int64    `json:"id,omitempty" datastore:"-"`
	Fields []string `json:"fields" datastore:"fields" validate:"required"`
	Op     string   `json:"op" datastore:"op" validate:"required,eq=add|eq=sub"`
}

func (w *Webhook) LoadKey(k *datastore.Key) error {
	w.ID = k.ID
	return nil
}

func (w *Webhook) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(w, ps)
}

func (w *Webhook) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(w)
}

type ComputationBody struct {
	ValueMap map[string]int64 `json:"values" validator:"required"`
}

type ComputationValues struct {
	Key   string
	Value int64
}

type Computation struct {
	ID        int64               `json:"id" datastore:"-"`
	WebhookID int64               `json:"webhookId" datastore:"webhookId"`
	Values    []ComputationValues `json:"values" datastore:"values"`
	Computed  bool                `json:"computed" datastore:"computed"`
	Result    *int64              `json:"result" datastore:"result"`
}

type PubSubPayload struct {
	ComputationId int64            `json:"computation_id"`
	Op            string           `json:"op"`
	Fields        []string         `json:"fields"`
	Values        map[string]int64 `json:"values"`
}
