package helloworld

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"net/http"
	"time"
)

var psClient *pubsub.Client
var psStoreResultTopic *pubsub.Topic

type fOp func (result int64, value int64) int64

type PubSubEvent struct {
	PubSubMessage struct {
		Data string `json:"data"`
	} `json:"message"`
}

type PubSubPayload struct {
	ComputationId int64            `json:"computation_id"`
	Op            string           `json:"op"`
	Fields        []string         `json:"fields"`
	Values        map[string]int64 `json:"values"`
}

type DataFlowPayload struct {
	ComputationId int64 `json:"computation_id"`
	Result int64 `json:"result"`
}

func init() {
	var err error
	ctx := context.Background()

	psClient, err = pubsub.NewClient(ctx, "paul-test-321412")
	if err != nil {
		log.Fatal(err)
	}
	psStoreResultTopic = psClient.Topic("result")
}

func add(result int64, value int64) int64 {
	return result + value
}

func sub(result int64, value int64) int64 {
	return result - value
}

func ProcessCompute(payload PubSubPayload) int64 {
	fs := map[string]fOp{
		"add": add,
		"sub": sub,
	}

	if f, ok := fs[payload.Op]; ok {
		result := payload.Values[payload.Fields[0]]

		for _, key := range payload.Fields[1:] {
			if value, ok := payload.Values[key]; ok {
				result = f(result, value)
			} else {
				panic(1)
			}
		}
		return result
	}
	panic(0)
}

func ComputeResult(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := PubSubEvent{}

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
		return
	}
	decodedData, _ := base64.StdEncoding.DecodeString(msg.PubSubMessage.Data)
	computationPayload := PubSubPayload{}
	if err := json.Unmarshal(decodedData, &computationPayload); err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
		return
	}
	newPayload, _ := json.Marshal(DataFlowPayload{
		ComputationId: computationPayload.ComputationId,
		Result:        ProcessCompute(computationPayload),
	})
	_ = psStoreResultTopic.Publish(ctx, &pubsub.Message{
		Data: newPayload,
	})
}