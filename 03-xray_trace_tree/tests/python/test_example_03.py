import logging
import json
import pathlib
import time
from unittest import TestCase

import boto3
import aws_ctk


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)


def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

class Example03(TestCase):
    stack_name: str = "cdk-example-sfnStack"
    stack_outputs: dict = read_cdk_outputs().get(stack_name, {}) 
    statemachine_arn: str = stack_outputs["StateMachineArn"]
    z: aws_ctk.AWSCtk = aws_ctk.AWSCtk()
    # patch sfn client to ensure trace is sampled
    sfn_client: boto3.client = z.patch_aws_client(boto3.client("stepfunctions"))

    def setUp(self):
        self.tracing_header = None

        response = self.sfn_client.start_execution(
            stateMachineArn=self.statemachine_arn,
            input=json.dumps({"waitMilliseconds": 1000}),
        )
        execution_arn = response["executionArn"]
        status = "RUNNING"
        while status == "RUNNING":
            res = self.sfn_client.describe_execution(
                executionArn=execution_arn,
            )
            status = res["status"]
            if not self.tracing_header:
                self.tracing_header = res["traceHeader"]


    def test_get_trace_tree(self):
        time.sleep(5)
        trace_tree = self.z.get_trace_tree(
            tracing_header=self.tracing_header,
        ).trace_tree

        self.assertEqual(len(trace_tree.paths), 3)
        self.assertEqual(
            [[seg.origin for seg in path] for path in trace_tree.paths],
            [
                ["AWS::StepFunctions::StateMachine", "AWS::Lambda"],
                ["AWS::StepFunctions::StateMachine", "AWS::Lambda"],
                ["AWS::StepFunctions::StateMachine", "AWS::SNS"],
            ]
        )
        
    def test_retry_get_trace_tree_until(self):
        def assertion(output: aws_ctk.GetTraceTreeOutput) -> None:
            tree = output.trace_tree
            self.assertEqual(len(tree.paths), 3)
            self.assertEqual(
                [[seg.origin for seg in path] for path in tree.paths],
                [
                    ["AWS::StepFunctions::StateMachine", "AWS::Lambda"],
                    ["AWS::StepFunctions::StateMachine", "AWS::Lambda"],
                    ["AWS::StepFunctions::StateMachine", "AWS::SNS"],
                ]
            )

        self.assertTrue(self.z.retry_get_trace_tree_until(
            tracing_header=self.tracing_header,
            assertion_fn=assertion,
            timeout_seconds=20,
        ))
