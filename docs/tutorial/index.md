---
title: Tutorial
description: Introduction to Zion
---

This tutorial introduces Zion by going through four examples. Each of them showcases one feature at a time.

For each example, we will execute the following steps:

1. Deploy System Under Test (SUT) with [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html){target="_blank"} or [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html)
2. Run the example test code with [pytest](https://docs.pytest.org/){target="_blank"}

## Requirements

* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html){target="_blank"} and [configured with your credentials](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-getting-started-set-up-credentials.html){target="_blank"}.
* [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html){target="_blank"} installed.
* [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html) installed.

## Getting started

Clone the examples:

```bash
git clone --single-branch --branch examples git@github.com:awslabs/aws-zion-private.git zion-examples
cd zion-examples
```

To run the Python (3.8+) examples:

```bash
python -m venv. venv
source .venv/bin/activate

pip install -r requirements.txt
```

## Retrieving information from a deployed CloudFormation Stack

This example shows how to use `get_stack_outputs` and `get_physical_id_from_stack` to retrieve information from a deployed CloudFormation Stack. They are useful if you deploy your stack directly with a CloudFormation template.

### System Under Test

We will use SAM CLI to deploy the SUT to CloudFormation. The SUT consists of one SQS Queue. After deploying the SUT, the Queue URL of the queue can be retrieved from both Physical ID and Outputs.

=== "01-cfn_utils/template.json"
    ```json
    {
        "AWSTemplateFormatVersion": "2010-09-09",
        "Description": "simple SQS template",
        "Resources": {
            "SQSQueue": {
                "Type": "AWS::SQS::Queue"
            }
        },
        "Outputs": {
            "QueueURL": {
                "Description": "URL of newly created SQS Queue",
                "Value": {
                    "Ref": "SQSQueue"
                }
            },
            "QueueURLFromGetAtt": {
                "Description": "Queue URL",
                "Value": {
                    "Fn::GetAtt": [
                        "SQSQueue",
                        "QueueUrl"
                    ]
                }
            },
            "QueueArn": {
                "Description": "Queue ARN",
                "Value": {
                    "Fn::GetAtt": [
                        "SQSQueue",
                        "Arn"
                    ]
                }
            }
        }
    }

    ```

To deploy the stack using AWS SAM CLI:

```bash
cd "01-cfn_utils"

sam deploy --stack-name example-01 --template ./template.json
```

### Test Code

#### Python

In the test code, we use both `get_stack_outputs` and `get_physical_id_from_stack` to get the Queue URL, then assert the values returned from both methods are equal.

=== "01-cfn_utils/test_example_01.py"
```python
import os
import zion

def test_zion_utils():
    stack_name = os.getenv("STACK_NAME", "example-01")
    region = os.getenv("AWS_REGION", "us-east-1")
    z = zion.Zion(region=region)

    outputs = z.get_stack_outputs(
        zion.GetStackOutputsParams(
            stack_name=stack_name,
            output_names=["QueueURL"],
        )
    ).outputs

    physical_id = z.get_physical_id_from_stack(
        zion.PhysicalIdFromStackParams(
            stack_name=stack_name,
            logical_resource_id="SQSQueue",
        )
    ).physical_id

    assert physical_id == outputs["QueueURL"]
```

To run the test code:

```bash
pytest test_example_01.py
```

## Testing EventBridge Event Bus with "Listener"

This example shows how to use a "Listener" to test a Rule on a given Event Bus. A "Listener" is a "Test Harness" that Zion helps you create for testing event delivery.

### System Under Test

In this example, we use AWS CDK to define the SUT. The SUT consists of these resources:

* an API Gateway Rest API (Entry Point)
* a Lambda Function (Producer)
* an Eventbridge Event Bus
* an Eventbridge Rule
* a Lambda Function (Consumer), as a target of the Rule

When the Rest API receives a request, it invokes the Producer. The Producer then sends an event to the Event Bus, which then delivers the event to Consumer acoording to the Rule.

We added some `CfnOutput` constructs to expose certain attributes from the SUT. These include:

* the name of the Event Bus
* the URL of the API endpoint
* the Eventbridge Rule name
* the Target ID in the Rule

These values will be used during the tests.

=== "02-eb_listener/cdk/lib/eb-stack.ts"
```typescript
import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as events from 'aws-cdk-lib/aws-events';
import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as path from 'path';

export class EbStack extends cdk.Stack {
    eventbus: events.EventBus | null = null;
    rule: events.Rule | null = null;
    target: targets.LambdaFunction | null = null;
    api: apigateway.RestApi | null = null;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        this.eventbus = new events.EventBus(this, 'EB');

        const producer = new lambda.Function(this, 'Producer', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'producerHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
            environment: {
                EVENTBUS_NAME: this.eventbus.eventBusName,
            },
            tracing: lambda.Tracing.ACTIVE,
        });
        this.eventbus.grantPutEventsTo(producer);

        this.api = new apigateway.RestApi(this, 'API', {
            deploy: true,
            deployOptions: {
                tracingEnabled: true,
            },
        });
        const resource = this.api.root.addResource('orders');
        const integration = new apigateway.LambdaIntegration(producer, {
            proxy: true,
        });
        resource.addMethod('POST', integration);

        const consumer = new lambda.Function(this, 'Consumer', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'consumerHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
            environment: {
                EVENTBUS_NAME: this.eventbus.eventBusName,
            },
            tracing: lambda.Tracing.ACTIVE,
        });

        this.rule = new events.Rule(this, 'ConsumerRule', {
            eventBus: this.eventbus,
            eventPattern: {
                source: ['com.hello-world.producer'],
                detailType: ['NewOrder'],
            },
        });
        this.target = new targets.LambdaFunction(consumer, {
            event: events.RuleTargetInput.fromEventPath('$.detail.customerId'),
        });
        this.rule.addTarget(this.target);

        this.output();
    }

    output() {
        if (this.eventbus) {
            new cdk.CfnOutput(this, 'EventBusName', {
                description: 'Event Bus Name',
                value: this.eventbus.eventBusName,
            });
        }

        if (this.rule) {
            new cdk.CfnOutput(this, 'RuleName', {
                description: 'Rule Name',
                value: this.rule.ruleName,
            });
            new cdk.CfnOutput(this, 'TargetId', {
                description: 'Target Id',
                value: 'Target0',
            });
        }
        if (this.api) {
            new cdk.CfnOutput(this, 'ApiEndpoint', {
                description: 'API Endpoint',
                value: this.api.urlForPath('/orders'),
            });
        }
    }
}
```

To deploy the SUT:

```bash
# navigate to the example dir
cd "02-eb_listener"

# install dependencies for building and deploying
npm install

# Deploy the stack using cdk, see package.json for definition of the command:
npm run deploy

```

After deploying, an output file `outputs.json` is created, with contents similar to below:

=== "outputs.json"
```json
{
    "example-ebStack": {
        "EventBusName": "examplestack01EB321ED36B",
        "ApiEndpoint": "https://xxxxxxxxx.execute-api.us-east-1.amazonaws.com/prod/orders",
        "APIEndpoint1793E782": "https://xxxxxxxxx.execute-api.us-east-1.amazonaws.com/prod/",
        "RuleName": "examplestack01EB321ED36B|example-stack-01-ConsumerRuleEE1F6314-12K2NOJQRM8A6",
        "TargetId": "Target0"
    }
}

```

### Test Code

#### Python

In the test code, we follow the "Arrage, Act, Assert" pattern. In Python, we do it by using [`unittest.TestCase`](https://docs.python.org/3/library/unittest.html#unittest.TestCase){target="_blank"}. We use `setUpClass` and `tearDownClass` to create and destroy Test Harnesses before and after individual tests respectively. Specifically:

* In `setUpClass`, we first call `remove_listeners` with `tag_filters` to destroy any previous orphaned listener. Then we call `add_listener` to create a listener by providing the Event Bus Name, the Rule Name, and the Target ID. Those values are retrieved from the "outputs.json" file. We also attach a tag to the listener so we can look it up more easily. The `add_listener` returns the listener ID. We keep the listener ID throughout the tests.
* In `tearDownClass`, we call `remove_listeners` to the listener created during `setUpClass`.

We have two tests `test_event_lands_at_eb` and `test_poll_events`, which showcase the `wait_until_event_matched` method and the `poll_events` method respectively:

* In `test_event_lands_at_eb`, we define a function `match_fn` to determine if a received event is matching expectation. We supply `match_fn` to the `wait_until_event_matched` method as an argument. The method will keep polling events from the listener until the given `match_fn` returns true or until timeout.
* In `test_poll_events`, we call the `poll_events` method. This method is a primitive method of `wait_until_event_matched`, i.e. it polls from the listener just once.

=== "02-eb_listener/tests/python/test_example_02.py"

```python
import logging
import json
import pathlib
from unittest import TestCase

import requests
import zion


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)


def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

class Example02(TestCase):
    stack_name: str = "cdk-example-ebStack"
    stack_outputs: dict = read_cdk_outputs().get(stack_name, {}) 
    z: zion.Zion = zion.Zion()

    @classmethod
    def setUpClass(cls) -> None:
        cls.event_bus_name = cls.stack_outputs["EventBusName"]
        cls.api_endpoint = cls.stack_outputs["ApiEndpoint"]
        cls.rule_name = cls.stack_outputs["RuleName"].split("|")[1]
        cls.target_id = cls.stack_outputs["TargetId"]

        # remote orphaned listeners from previous test runs (if any)
        cls.z.remove_listeners(
            zion.RemoveListenersParams(
                tag_filters=[
                    zion.RemoveListeners_TagFilter(
                        key="stage",
                        values=["example02"],
                    )
                ]
            )
        )

        # create listener
        listener_id = cls.z.add_listener(
            zion.AddEbListenerParams(
                event_bus_name=cls.event_bus_name,
                rule_name=cls.rule_name,
                target_id=cls.target_id,
                tags={"stage": "example02"},
            )
        ).id
        cls.listeners = [listener_id]
        LOG.debug("created listeners: %s", cls.listeners)
        super().setUpClass()

    @classmethod
    def tearDownClass(cls) -> None:
        cls.z.remove_listeners(
            zion.RemoveListenersParams(
                ids=cls.listeners,
            )
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
                zion.WaitUntilEventMatchedParams(
                    listener_id=self.listeners[0],
                    condition=match_fn,
                )
            )
        )

    def test_poll_events(self):
        customer_id = "def456"
        requests.post(self.api_endpoint, params={"customerId": customer_id})

        received = self.z.poll_events(
            zion.PollEventsParams(
                listener_id=self.listeners[0],
                wait_time_seconds=5,
                max_number_of_messages=10,
            )
        ).events
        LOG.debug("received: %s", received)
        self.assertGreaterEqual(len(received), 1)
        self.assertEqual(json.loads(received[0]), customer_id)

```

To run the test code:

```bash
pytest tests/python/test_example_02.py
```

## Testing with X-Ray Traces

This example shows how to test with X-Ray Traces. If you have X-Ray instrumented throughout your application, X-Ray Traces provide a good amount of details for you to inspect for testing purpose. Zion helps you fetch traces and parse them into easily-queryable objects for inspectation. For example, you can easily verify if a trace hit an expected sequeunce of AWS resources.

### System Under Test

In this example, we use AWS CDK to define the SUT. The SUT consists of one StepFunction State Machine.

We added some `CfnOutput` constructs to expose certain attributes from the SUT. These include:

* the name of the State Machine
* the ARN of the State Machine

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

In the test code, we first implement the `setUp` method to start a State Machine execution, and wait for the execution to complete. We keep the tracing header of the execution.

We have two tests `test_get_trace_tree` and `test_retry_get_trace_tree_until`, which showcase the `get_trace_tree` method and the `retry_get_trace_tree_until` method:

* In `test_get_trace_tree`, we added a sleep of 5 seconds to wait for the trace to be fetchable. The `get_trace_tree` uses the fetched Trace to build and return a Trace Tree. Nodes of the tree are segments or subsegments of the trace. The root is the starting point of the trace. The method returns the root segment and also all the paths in the tree. Each path is a sequence of nodes from root to leaf. In the test, we extract the `origin` attribute of each node of each path to assert if the expected sequence of AWS resources were invoked. 
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
            zion.GetTraceTreeParams(
                tracing_header=self.tracing_header,
            )
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
        def condition(output: zion.GetTraceTreeOutput) -> bool:
            tree = output.trace_tree
            try:
                self.assertEqual(len(tree.paths), 3)
                self.assertEqual(
                    [[seg.origin for seg in path] for path in tree.paths],
                    [
                        ["AWS::StepFunctions::StateMachine", "AWS::Lambda"],
                        ["AWS::StepFunctions::StateMachine", "AWS::Lambda"],
                        ["AWS::StepFunctions::StateMachine", "AWS::SNS"],
                    ]
                )
                return True
            except AssertionError:
                return False

        self.assertTrue(self.z.retry_get_trace_tree_until(
            zion.RetryGetTraceTreeUntilParams(
                tracing_header=self.tracing_header,
                condition=condition,
                timeout_seconds=20,
            )
        ))

```

To run the test code:

```bash
pytest tests/python/test_example_03.py
```

## Generate Mock Events

Zion provides the capability for you to generate mock events from a schema stored in [Amazon Eventbridge schema registries](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-schema-registry.html){target="_blank"}. This allows you to generate a mock event and invoke any consumer (such as Lambda Function, StepFunction State Machine) with the generated event.

### System Under Test

In this example, we use AWS CDK to define the SUT. The SUT consists of one Schema Registry, one Schema and one Lambda Function.

We added some `CfnOutput` constructs to expose certain attributes from the SUT. These include:

* the name of the Lambda Function
* the name of the Schema Registry
* the name of the Schema

These values will be used during the tests.

=== "04-event_generation/cdk/lib/schema-stack.ts"
```typescript
import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as eventschemas from 'aws-cdk-lib/aws-eventschemas';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as path from 'path';

export class SchemaStack extends cdk.Stack {
    registry: eventschemas.CfnRegistry | null = null;
    schema: eventschemas.CfnSchema | null = null;
    lambdaFunction: lambda.Function | null = null;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        this.registry = new eventschemas.CfnRegistry(this, 'MyRegistry', {});
        this.schema = new eventschemas.CfnSchema(this, 'MySchema', {
            registryName: this.registry.attrRegistryName,
            type: 'OpenApi3',
            content: JSON.stringify({
                openapi: '3.0.0',
                info: {
                    version: '1.0.0',
                    title: 'my-event',
                },
                paths: {},
                components: {
                    schemas: {
                        MyEvent: {
                            type: 'object',
                            properties: {
                                customerId: {
                                    type: 'string',
                                },
                                datetime: {
                                    type: 'string',
                                    format: 'date-time',
                                },
                                membershipType: {
                                    type: 'string',
                                    enum: ['A', 'B', 'C'],
                                },
                                address: {
                                    type: 'string',
                                },
                                orderItems: {
                                    type: 'array',
                                    items: {
                                        $ref: '#/components/schemas/Item',
                                    },
                                },
                            },
                        },
                        Item: {
                            type: 'object',
                            properties: {
                                sku: {
                                    type: 'string',
                                },
                                unitPrice: {
                                    type: 'number',
                                },
                                count: {
                                    type: 'integer',
                                },
                            },
                        },
                    },
                },
            }),
        });

        this.lambdaFunction = new lambda.Function(this, 'Calculator', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'calculatorHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
        });

        // outputs
        new cdk.CfnOutput(this, 'CalculatorFunction', {
            description: 'Lambda Function Name',
            value: this.lambdaFunction.functionName,
        });
        new cdk.CfnOutput(this, 'RegistryName', {
            value: this.registry.attrRegistryName,
        });
        new cdk.CfnOutput(this, 'SchemaName', {
            value: this.schema.attrSchemaName,
        });
    }
}
```

To deploy the SUT:

```bash
# navigate to the example dir
cd "04-event_generation"

# Deploy the stack using cdk, see package.json for definition of the command:
npm run deploy
```

After deploying, an output file `outputs.json` is created, with contents similar to below:

=== "outputs.json"
```json
{
  "example-schemaStack": {
    "CalculatorFunction": "cdk-example-schemaStack-CalculatorBxxxxF40-5SYJsAlTscGC",
    "RegistryName": "MyRegsitry-xx5rNdAMGJL1",
    "SchemaName": "MySchema-xxKd3I1NbYAu"
  }
}
```

### Test Code

#### Python

In the test, we use three tests, `test_generate_barebone_event`, `test_generate_contextful_event` and `test_generate_eventbridge_event`, to demostrate how you can generate mock events:

In `test_generate_barebone_event`, we call `generate_mock_event` by providing only `registry_name`, `schema_name` and `event_ref`. This gives you a "barebone" event:

```json
{
  "address": "",
  "customerId": "",
  "datetime": "2023-10-18T21:08:04.782196-07:00",
  "membershipType": "A",
  "orderItems": []
}
```

As shown in `test_generate_contextful_event`, you can supply contexts to enrich the generated event. We defined a function `apply_context` which populates the `customerId` field and the `address` field, and also add five items into `orderItems`. We then supply this function into the `contexts` argument in the `generate_mock_event` call. Note that `contexts` accepts a list of functions, meaning that you can apply multiple contexts. The generated events looks like:

```json
{
  "address": "99 Some Street",
  "customerId": "8e9bf525-168c-47ad-96e6-507dd4a15ba5",
  "datetime": "2023-10-18T21:08:05.31715-07:00",
  "membershipType": "A",
  "orderItems": [
    {
      "unitPrice": 2,
      "count": 1
    },
    {
      "unitPrice": 4,
      "count": 2
    },
    {
      "unitPrice": 6,
      "count": 3
    },
    {
      "unitPrice": 8,
      "count": 4
    },
    {
      "unitPrice": 10,
      "count": 5
    }
  ]
}
```

In the same test, we then use the generated event as payload to invoke the Lambda Function, and assert if the return from the invocation equals to the expected value.

As shown in `test_generate_eventbridge_event`, if you are generating Event Bridge events, Zion provides `zion.context_generation.eventbridge_event_context` for you to enrich a barebone Event Bridge event.

=== "04-event_generation/tests/python/test_example_04.py"
```python
import logging
import json
import pathlib
import uuid

import boto3
import zion
from zion.context_generation import eventbridge_event_context

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
lambda_client = boto3.client("lambda")

def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

def test_generate_barebone_event():
    stack_name = "cdk-example-schemaStack"
    stack_outputs = read_cdk_outputs().get(stack_name, {})
    registry_name = stack_outputs["RegistryName"]
    schema_name = stack_outputs["SchemaName"]
    z = zion.Zion()
    barebone_event = z.generate_mock_event(
        zion.GenerateMockEventParams(
            registry_name=registry_name,
            schema_name=schema_name,
            event_ref="MyEvent",
        )
    ).event
    for key in ["address", "customerId", "datetime", "membershipType", "orderItems"]:
        assert key in barebone_event
    assert barebone_event["address"] == ""
    assert barebone_event["customerId"] == ""
    assert barebone_event["orderItems"] == []
    
def test_generate_contextful_event():
    stack_name = "cdk-example-schemaStack"
    stack_outputs = read_cdk_outputs().get(stack_name, {})
    registry_name = stack_outputs["RegistryName"]
    schema_name = stack_outputs["SchemaName"]
    function_name = stack_outputs["CalculatorFunction"]
    z = zion.Zion()
    
    def apply_context(event: dict) -> dict:
        event["customerId"] = str(uuid.uuid4())
        event["address"] = "99 Some Street"
        for i in range(5):
            item = {
                "unitPrice": (i + 1) * 2,
                "count": i + 1, 
            }
            event["orderItems"].append(item)
        return event
        
    mock_event = z.generate_mock_event(
        zion.GenerateMockEventParams(
            registry_name=registry_name,
            schema_name=schema_name,
            event_ref="MyEvent",
            contexts=[apply_context],
        )
    ).event
    for key in ["address", "customerId", "datetime", "membershipType", "orderItems"]:
        assert key in mock_event
    assert mock_event["customerId"] != ""
    assert mock_event["address"] == "99 Some Street"
    assert len(mock_event["orderItems"]) > 0

    response = lambda_client.invoke(
        FunctionName=function_name,
        Payload=bytes(json.dumps(mock_event), encoding="utf-8"),
    )
    result = int(response['Payload'].read())
    assert result == 110

def test_generate_eventbridge_event():
    z = zion.Zion()

    mock_eb_event = z.generate_mock_event(
        zion.GenerateMockEventParams(
            registry_name="aws.events",
            schema_name="aws.autoscaling@EC2InstanceLaunchSuccessful",
            schema_version="2",
            event_ref="AWSEvent",
            contexts=[eventbridge_event_context],
        )
    ).event
    LOG.debug(mock_eb_event)
    for key in ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]:
        assert key in mock_eb_event
    assert mock_eb_event["id"] != ""
    assert mock_eb_event["account"] != ""
    assert mock_eb_event["time"] != ""

```

To run the test code:

```bash
pytest tests/python/test_example_04.py
```
