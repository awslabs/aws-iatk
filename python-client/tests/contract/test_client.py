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
        if len(signature.parameters) != 2 or "params" not in signature.parameters:
            raise NotAClientMethodException(
                f"{name} is not a client method of zion.Zion"
            )
        self.param_cls = signature.parameters["params"].annotation
        self.returns_cls = signature.return_annotation

    def is_rpc_method(self):
        return hasattr(self.param_cls, "_rpc_method")

    @property
    def rpc_method(self):
        if not self.is_rpc_method():
            return None
        return getattr(self.param_cls, "_rpc_method")

    @property
    def param_annotations(self):
        return [
            n
            for n in inspect.get_annotations(self.param_cls).keys()
            if not (n.startswith("_") or n == "jsonrpc_dumps")
        ]

    @property
    def returns_annotations(self):
        return [n for n in inspect.get_annotations(self.returns_cls).keys()]


def to_snake_case(name):
    return "".join(["_" + c.lower() if c.isupper() else c for c in name]).lstrip("_")


class ClientContractTest(TestCase):
    specs: dict

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
        rpc_methods = [m for m in self.client_methods.values() if m.is_rpc_method()]
        self.assertEqual(len(self.specs["methods"]), len(rpc_methods))
        for m in rpc_methods:
            self.assertIn(m.rpc_method, self.specs["methods"])

    def test_param_properties(self):
        for name, method in self.client_methods.items():
            if not method.is_rpc_method():
                continue
            params_spec = (
                self.specs["methods"].get(method.rpc_method, {}).get("parameters")
            )
            spec_params = set(
                [to_snake_case(k) for k in params_spec["properties"].keys()]
            )
            self.assertEqual(
                spec_params - set(["region", "profile"]),
                set(method.param_annotations),
                f"method: {name}",
            )

    def test_returns_properties(self):
        for name, method in self.client_methods.items():
            if not method.is_rpc_method():
                continue
            if name in [
                "get_physical_id_from_stack",
                "get_stack_outputs",
                "remove_listeners",
                "poll_events",
            ]:
                # NOTE: skipping for these methods since the Output cls does some form of transform
                continue
            returns_spec = (
                self.specs["methods"].get(method.rpc_method, {}).get("returns")
            )
            if "properties" not in returns_spec:
                continue
            spec_returns = set(
                [to_snake_case(k) for k in returns_spec["properties"].keys()]
            )
            self.assertEqual(
                spec_returns, set(method.returns_annotations), f"method: {name}"
            )
