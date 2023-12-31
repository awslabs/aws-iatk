---
title: Testing EventBridge Event Bus with "Listener"
description: Example to showcase how to use Listener to test a Rule on a given Event Bus
---

In this example, we use a "Listener" to test a rule on an Amazon EventBridge event bus. A "Listener" is a "Test Harness" that AWS IATK helps you create for testing event delivery.

### System Under Test

For this example, we use AWS CDK to define the System Under Test (SUT). The SUT consists of these resources:

* An Amazon API Gateway REST API (Entry Point).
* An AWS Lambda function (Producer).
* An EventBridge event bus.
* An EventBridge rule.
* A Lambda function (Consumer), as a target of the rule.

When the REST API receives a request, it invokes the producer. The producer then sends an event to the event bus, which then delivers the event to the consumer according to the rule.

We added some `CfnOutput` constructs to expose certain attributes from the System Under Test (SUT). These include:

* The name of the event bus.
* The URL of the API endpoint.
* The EventBridge rule name.
* The target ID in the rule.

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

After deploying, the output file `outputs.json` is created, with contents similar to below:

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

In the test code, we follow the "Arrage, Act, Assert" pattern. In Python, we do this by using [`unittest.TestCase`](https://docs.python.org/3/library/unittest.html#unittest.TestCase){target="_blank"}. We use `setUpClass` and `tearDownClass` to create and destroy Test Harnesses before and after individual tests respectively. Specifically:

* In `setUpClass`, we first call `remove_listeners` with `tag_filters` to shut down any previous orphaned listener. Then we call `add_listener` to create a listener by providing the event bus name, the rule name, and the target ID. Those values are retrieved from the "outputs.json" file. We also attach a tag to the listener so we can look it up more easily. The `add_listener` returns the listener ID. We keep the listener ID throughout the tests.
* In `tearDownClass`, we call `remove_listeners` to shut down the listener created during `setUpClass`.

We have two tests, `test_event_lands_at_eb` and `test_poll_events`, which showcase the `wait_until_event_matched` method and the `poll_events` method respectively:

* In `test_event_lands_at_eb`, we define a function `assert_fn` to determine if a received event is matching expectation. We supply `assert_fn` to the `wait_until_event_matched` method as an argument. The method will keep polling events from the listener until the given `assert_fn` succeeds or until timeout.
* In `test_poll_events`, we call the `poll_events` method. This method is a primitive method of `wait_until_event_matched`, i.e. it polls from the listener just once.

=== "02-eb_listener/tests/python/test_example_02.py"

```python
import logging
import json
import pathlib
from unittest import TestCase

import requests
import aws_iatk


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)


def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

class Example02(TestCase):
    stack_name: str = "cdk-example-ebStack"
    stack_outputs: dict = read_cdk_outputs().get(stack_name, {}) 
    iatk: aws_iatk.AwsIatk = aws_iatk.AwsIatk()

    @classmethod
    def setUpClass(cls) -> None:
        cls.event_bus_name = cls.stack_outputs["EventBusName"]
        cls.api_endpoint = cls.stack_outputs["ApiEndpoint"]
        cls.rule_name = cls.stack_outputs["RuleName"].split("|")[1]
        cls.target_id = cls.stack_outputs["TargetId"]

        # remote orphaned listeners from previous test runs (if any)
        cls.iatk.remove_listeners(
            tag_filters=[
                aws_iatk.RemoveListeners_TagFilter(
                    key="stage",
                    values=["example02"],
                )
            ]
        )

        # create listener
        listener_id = cls.iatk.add_listener(
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

        def assert_fn(received: str) -> None:
            received = json.loads(received)
            LOG.debug("received: %s", received)
            assert received == customer_id

        self.assertTrue(
            self.iatk.wait_until_event_matched(
                listener_id=self.listeners[0],
                assertion_fn=assert_fn,
            )
        )

    def test_poll_events(self):
        customer_id = "def456"
        requests.post(self.api_endpoint, params={"customerId": customer_id})

        received = self.iatk.poll_events(
            listener_id=self.listeners[0],
            wait_time_seconds=5,
            max_number_of_messages=10,
        ).events
        LOG.debug("received: %s", received)
        self.assertGreaterEqual(len(received), 1)
        self.assertEqual(json.loads(received[0]), customer_id)

```

To run the test code:

```bash
pytest tests/python/test_example_02.py
```
