import argparse
import json
import logging
from abc import ABC

import apache_beam as beam
from apache_beam.io import ReadFromPubSub
from apache_beam.options.pipeline_options import PipelineOptions
from google.cloud import datastore

project_id = "paul-test-321412"
subscription_id = "result-sub"
subscription = "projects/{}/subscriptions/{}".format(project_id, subscription_id)


class EditEntityDoFn(beam.DoFn, ABC):
    def __init__(self, kind, *unused_args, **unused_kwargs):
        super().__init__(*unused_args, **unused_kwargs)
        self.kind = kind

    def process(self, element, **kwargs):
        client = datastore.Client(project=project_id)

        cmp_id = element["computation_id"]
        key = client.key(self.kind, cmp_id)
        entity = client.get(key)

        if entity is None:
            return

        entity["result"] = element["result"]
        entity["computed"] = True
        client.put(entity)


def parse_pub_sub(element):
    try:
        payload = json.loads(element)
    except json.JSONDecodeError:
        logging.error(f"invalid_json_payload")
        return None

    if len(payload) > 2 | \
            ("computation_id" in payload is False) | \
            ("result" in payload is False):
        logging.error(f"invalid_json_payload")
        return None
    if (type(payload["computation_id"]) != int) | (type(payload["result"]) != int):
        return None
    return payload


def run(args):
    options = PipelineOptions(args, streaming=True, save_main_session=True)

    with beam.Pipeline(options=options) as p:
        (
                p
                | 'Read from PS' >> ReadFromPubSub(subscription=subscription).with_output_types(bytes)
                | 'Decode PS message' >> beam.Map(lambda x: x.decode("utf-8"))
                | 'Get json' >> beam.Map(parse_pub_sub)
                | 'Get and edit entity' >> beam.ParDo(EditEntityDoFn(kind="Computation"))
        )


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    _, args = parser.parse_known_args()
    logging.getLogger().setLevel(logging.INFO)
    run(args)
