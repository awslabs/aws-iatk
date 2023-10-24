---
title: Generate Mock Events
description: Example to showcase how to generate mock events
---

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

# install dependencies for building and deploying
npm install

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
        registry_name=registry_name,
        schema_name=schema_name,
        event_ref="MyEvent",
    ).event
    LOG.debug(json.dumps(barebone_event, indent=2))
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
        registry_name=registry_name,
        schema_name=schema_name,
        event_ref="MyEvent",
        contexts=[apply_context],
    ).event
    LOG.debug(json.dumps(mock_event, indent=2))
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
        registry_name="aws.events",
        schema_name="aws.autoscaling@EC2InstanceLaunchSuccessful",
        schema_version="2",
        event_ref="AWSEvent",
        contexts=[eventbridge_event_context],
    ).event
    LOG.debug(json.dumps(mock_eb_event, indent=2))
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
