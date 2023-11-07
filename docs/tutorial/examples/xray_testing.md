---
title: Testing with X-Ray traces
description: Example to showcase how to test with X-Ray traces
---

This example shows how to test with AWS X-Ray traces. When implemented throughout your application, X-Ray traces provides a good amount of detail that you can inspect for testing purposes. AWS CTK helps you fetch traces and parse them into objects that can easily be queried for inspection. For example, you can easily verify if a trace hits an expected sequence of AWS resources.

### System Under Test (SUT)

In this example, we use AWS CDK to define the SUT. The SUT consists of one AWS Step Functions State Machine.

We added some `CfnOutput` constructs to expose certain attributes from the SUT. These include:

* The name of the state machine.
* The ARN of the state machine.

These values will be used during the tests.

=== "03-xray_trace_tree/cdk/lib/sfn-stack.ts"
```typescript
import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as sfn from 'aws-cdk-lib/aws-stepfunctions';
import * as tasks from 'aws-cdk-lib/aws-stepfunctions-tasks';
import * as sns from 'aws-cdk-lib/aws-sns';

export class SfnStack extends cdk.Stack {
    statemachine: sfn.StateMachine | null = null;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const convertToSeconds = new tasks.EvaluateExpression(this, 'Convert to seconds', {
            expression: '$.waitMilliseconds / 1000',
            resultPath: '$.waitSeconds',
        });

        const createMessage = new tasks.EvaluateExpression(this, 'Create message', {
            // Note: this is a string inside a string.
            expression: '`Now waiting ${$.waitSeconds} seconds...`',
            runtime: lambda.Runtime.NODEJS_LATEST,
            resultPath: '$.message',
        });

        const publishMessage = new tasks.SnsPublish(this, 'Publish message', {
            topic: new sns.Topic(this, 'cool-topic'),
            message: sfn.TaskInput.fromJsonPathAt('$.message'),
            resultPath: '$.sns',
        });

        const wait = new sfn.Wait(this, 'Wait', {
            time: sfn.WaitTime.secondsPath('$.waitSeconds'),
        });

        const definition = convertToSeconds.next(createMessage).next(publishMessage).next(wait);
        this.statemachine = new sfn.StateMachine(this, 'MyStateMachine', {
            definitionBody: sfn.DefinitionBody.fromChainable(definition),
            tracingEnabled: true,
        });

        new cdk.CfnOutput(this, 'StateMachineArn', {
            description: 'State Machine ARN',
            value: this.statemachine.stateMachineArn,
        });
        new cdk.CfnOutput(this, 'StateMachineName', {
            description: 'State Machine Name',
            value: this.statemachine.stateMachineName,
        });
    }
}

```

To deploy the SUT:

```bash
# navigate to the example dir
cd "03-xray_trace_tree"

# install dependencies for building and deploying
npm install

# Deploy the stack using cdk, see package.json for definition of the command:
npm run deploy

```

After deploying, an output file `outputs.json` is created, with contents similar to below:

=== "outputs.json"
```json
{
  "example-sfnStack": {
    "StateMachineArn": "arn:aws:states:us-east-1:123456789012:stateMachine:MyStateMachine6C968CA5-Ybusf26S5Oir",
    "StateMachineName": "MyStateMachine6C968CA5-Ybusf26S5Oir"
  }
}
```

### Test Code

#### Python

In the test code, we first implement the `setUp` method to start a state machine execution, and wait for the execution to complete. We keep the tracing header of the execution.

We have two tests, `test_get_trace_tree` and `test_retry_get_trace_tree_until`, which showcase the `get_trace_tree` method and the `retry_get_trace_tree_until` method:

* In `test_get_trace_tree`, we added a sleep of 5 seconds to wait for the trace to be fetchable. The `get_trace_tree` uses the fetched trace to build and return a trace tree. Nodes of the tree are segments or subsegments of the trace. The root is the starting point of the trace. The method returns the root segment and also all the paths in the tree. Each path is a sequence of nodes from root to leaf. In the test, we extract the `origin` attribute of each node of each path to assert if the expected sequence of AWS resources were invoked. 
* In `test_retry_get_trace_tree_until`, no sleep is needed as the `retry_get_trace_tree_until` method handles the latency issue by retrying fetching the trace with exponential backoff. In this test, we also define function called `condition` to do assertion on the returned Trace Tree. We supply `condition` to the `retry_get_trace_tree_until` as a stopping condition.

=== "03-xray_trace_tree/tests/python/test_example_03.py"
```python
import logging
import json
import pathlib
import time
from unittest import TestCase

import boto3
import zion


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
    z: zion.Zion = zion.Zion()
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
        def assertion(output: zion.GetTraceTreeOutput) -> None:
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

```

To run the test code:

```bash
pytest tests/python/test_example_03.py
```
