import logging
import json
import pathlib
from unittest import TestCase

import requests
import aws_ctk


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)


def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

class Example02(TestCase):
    stack_name: str = "cdk-example-ebStack"
    stack_outputs: dict = read_cdk_outputs().get(stack_name, {}) 
    z: aws_ctk.AWSCtk = aws_ctk.AWSCtk()

    @classmethod
    def setUpClass(cls) -> None:
        cls.event_bus_name = cls.stack_outputs["EventBusName"]
        cls.api_endpoint = cls.stack_outputs["ApiEndpoint"]
        cls.rule_name = cls.stack_outputs["RuleName"].split("|")[1]
        cls.target_id = cls.stack_outputs["TargetId"]

        # remote orphaned listeners from previous test runs (if any)
        cls.z.remove_listeners(
            tag_filters=[
                aws_ctk.RemoveListeners_TagFilter(
                    key="stage",
                    values=["example02"],
                )
            ]
        )

        # create listener
        listener_id = cls.z.add_listener(
            event_bus_name=cls.event_bus_name,
            rule_name=cls.rule_name,
            target_id=cls.target_id,
            tags={"stage": "example02"},
        ).id
        cls.listeners = [listener_id]
        LOG.debug("created listeners: %s", cls.listeners)
        super().setUpClass()

    @classmethod
    def tearDownClass(cls) -> None:
        cls.z.remove_listeners(
            ids=cls.listeners,
        )
        LOG.debug("destroyed listeners: %s", cls.listeners)
        super().tearDownClass()
            
    def test_event_lands_at_eb(self):
        customer_id = "abc123"
        requests.post(self.api_endpoint, params={"customerId": customer_id})

        def match_fn(received: str) -> bool:
            received = json.loads(received)
            LOG.debug("received: %s", received)
            return received == customer_id

        self.assertTrue(
            self.z.wait_until_event_matched(
                listener_id=self.listeners[0],
                condition=match_fn,
            )
        )

    def test_poll_events(self):
        customer_id = "def456"
        requests.post(self.api_endpoint, params={"customerId": customer_id})

        received = self.z.poll_events(
            listener_id=self.listeners[0],
            wait_time_seconds=5,
            max_number_of_messages=10,
        ).events
        LOG.debug("received: %s", received)
        self.assertGreaterEqual(len(received), 1)
        self.assertEqual(json.loads(received[0]), customer_id)
