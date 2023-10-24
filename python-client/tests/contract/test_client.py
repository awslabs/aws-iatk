# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Contract testing between client and RPC Specs
"""

import json
import logging
import inspect
from pathlib import Path
from unittest import TestCase

import zion

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
RPC_SPECS_LOC = "../../../schema/rpc-specs.json"


class NotAClientMethodException(Exception):
    pass


class ClientMethod:
    def __init__(self, name: str):
        obj = getattr(zion.Zion, name)
        signature = inspect.signature(obj)
        if name in ["patch_aws_client", "retry_until", "retry_get_trace_tree_until", "wait_until_event_matched"]:
            raise NotAClientMethodException(
                f"{name} is not a client method of zion.Zion"
            )
        self.params = signature.parameters.keys() - ['self']
        self.returns_cls = signature.return_annotation
        self.rpc_method = name

    @property
    def returns_annotations(self):
        return [n for n in inspect.get_annotations(self.returns_cls).keys()]


def to_snake_case(name):
    return "".join(["_" + c.lower() if c.isupper() else c for c in name]).lstrip("_")


class ClientContractTest(TestCase):
    specs: dict
    method_map: dict = {
        "add_listener": "test_harness.eventbridge.add_listener",
        "generate_mock_event": "mock.generate_barebone_event",
        "get_physical_id_from_stack": "get_physical_id",
        "poll_events": "test_harness.eventbridge.poll_events",
        "remove_listeners": "test_harness.eventbridge.remove_listeners"
    }

    @classmethod
    def setUpClass(cls) -> None:
        loc = Path(__file__).parent / RPC_SPECS_LOC
        with open(loc) as f:
            cls.specs = json.loads(f.read())

        cls.client_methods = {}
        for name in dir(zion.Zion):
            if callable(getattr(zion.Zion, name)) and not name.startswith("_"):
                try:
                    LOG.debug("name: %s", name)
                    cls.client_methods[name] = ClientMethod(name)
                except (NotAClientMethodException, TypeError, ValueError):
                    continue

    def test_all_rpc_methods_defined(self):
        rpc_methods = [m for m in self.client_methods.values()]
        for m in rpc_methods:
            self.assertIn(self.method_map.get(m.rpc_method, m.rpc_method), self.specs["methods"])
        self.assertEqual(len(self.specs["methods"]), len(rpc_methods))

    def test_param_properties(self):
        for name, method in self.client_methods.items():
            resolved_method = self.method_map.get(method.rpc_method, method.rpc_method)
            params_spec = (
                self.specs["methods"].get(resolved_method, {}).get("parameters")
            )
            spec_params = set(
                [to_snake_case(k) for k in params_spec["properties"].keys()]
            )
            
            # is this a client side only not in the rpc method sig
            if name == "generate_mock_event":
                spec_params.add("contexts") 
            
            self.assertEqual(
                spec_params - set(["region", "profile"]),
                set(method.params),
                f"method: {name}",
            )

    def test_returns_properties(self):
        for name, method in self.client_methods.items():
            if name in [
                "get_physical_id_from_stack",
                "get_stack_outputs",
                "remove_listeners",
                "poll_events",
                "get_trace_tree"
            ]:
                # NOTE: skipping for these methods since the Output cls does some form of transform
                continue
            
            resolved_method = self.method_map.get(method.rpc_method, method.rpc_method)
            returns_spec = (
                self.specs["methods"].get(resolved_method, {}).get("returns")
            )
            if "properties" not in returns_spec:
                continue
            spec_returns = set(
                [to_snake_case(k) for k in returns_spec["properties"].keys()]
            )
            self.assertEqual(
                spec_returns, set(method.returns_annotations), f"method: {name}"
            )
