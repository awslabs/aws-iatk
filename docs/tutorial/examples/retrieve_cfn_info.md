---
title: Retrieving information from a deployed CloudFormation Stack
description: Example to showcase how to retrieve info from a CloudFormation Stack
---

This example shows how to use `get_stack_outputs` and `get_physical_id_from_stack` to retrieve information from a deployed AWS CloudFormation Stack. They are useful if you deploy your stack directly with a CloudFormation template.

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
        stack_name=stack_name,
        output_names=["QueueURL"],
    ).outputs

    physical_id = z.get_physical_id_from_stack(
        stack_name=stack_name,
        logical_resource_id="SQSQueue",
    ).physical_id

    assert physical_id == outputs["QueueURL"]
    
```

To run the test code:

```bash
pytest test_example_01.py
```
