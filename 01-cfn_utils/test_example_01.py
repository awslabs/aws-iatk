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
    