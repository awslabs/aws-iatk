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
git clone --single-branch --branch examples git@github.com:awslabs/aws-zion-private.git
```

To run the Python (3.8+) examples:

```bash
python -m venv. venv
source .venv/bin/activate

pip install -r requirements.txt
```

## Retrieving information from a deployed stack

This example shows how to use `get_stack_outputs` and `get_physical_id_from_stack` to retrieve information from a deployed stack. They are useful if you deploy your stack directly with a CloudFormation template.

We will use SAM CLI to deploy the SUT to CloudFormation. For Python, we will use `pytest` to run the test code.

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
The SUT consists of only one SQS Queue. After deploying the SUT, the Queue URL of the queue can be retrieved from both Physical ID and Outputs. In the test code, we use both `get_stack_outputs` and `get_physical_id_from_stack` to get the Queue URL, and then assert the values returned from both methods are equal.

To deploy the stack using AWS SAM CLI:
```bash
cd "01-cfn_utils"

sam deploy --stack-name example-01 --template ./template.json
```

To run the test code:

=== "Python"
   ```bash
   pytest test_example_01.py
   ```

## Testing EventBridge Event Bus with "Listener"

This example shows how to use a "Listener" to test a Rule on a given Event Bus. A "Listener" is a "Test Harness" that Zion helps you create for testing event delivery. 

In this example, we use AWS CDK to define the SUT. The SUT consists of these resources:
- an API Gateway Rest API (Entry Point)
- a Lambda Function (Producer)
- an Eventbridge Event Bus
- an Eventbridge Rule
- a Lambda Function (Consumer), as a target of the Rule

=== "