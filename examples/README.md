# Examples

Below lists the examples to showcase how you can use Zion to write integration against the cloud more easily.

To run the examples in Python (3.7+):
```bash
# navigate into the examples dir from project root
cd examples

# setup venv
python -m venv .venv
source .venv/bin/activate

# install dependencies
pip install -r requirements.txt
```

## Example01 - retrieving information from a deployed stack

This example shows how to use `get_stack_outputs` and `get_physical_id_from_stack` to retrieve information from a deployed stack. They are useful if you deploy your stack directly with a CloudFormation template.

We will use SAM CLI to deploy a [stack](./example01/template.json) to CloudFormation. For Python, we will use `pytest` to run the [test code](./example01/test_example_01.py).

To setup SAM CLI, see [here](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)

To run the example:

```bash
# navigate into the example01 dir
cd example01

# To deploy the stack under test using SAM CLI:
sam deploy --stack-name example-01 --template ./template.json

# Run the Python example:
pytest test_example_01.py
```

To clean up the stack after running the example:

```bash
sam destroy --stack-name example-01
```

## Example02 - Example04

To run example02 - example04, we will use CDK to define and deploy the stacks under test.

To setup CDK, see [here](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html)

To deploy the stacks:

```bash
# navigate to the example02-04 dir
cd example02-04

# install dependencies for building and deploying
npm install

# Deploy the stack using cdk, see package.json for definition of the command:
npm run deploy

```

Note that, after deploy completes, an output file `outputs.json` is created, with contents similar to below:

```json
{
  "example-stack-01": {
    "EventBusName": "examplestack01EB321ED36B",
    "ApiEndpoint": "https://xxxxxxxxx.execute-api.us-east-1.amazonaws.com/prod/orders",
    "APIEndpoint1793E782": "https://xxxxxxxxx.execute-api.us-east-1.amazonaws.com/prod/",
    "RuleName": "examplestack01EB321ED36B|example-stack-01-ConsumerRuleEE1F6314-12K2NOJQRM8A6",
    "TargetId": "Target0"
  },
  "example-stack-02": {
    "StateMachineArn": "arn:aws:states:us-east-1:123456789012:stateMachine:MyStateMachine6C968CA5-Ybusf26S5Oir",
    "StateMachineName": "MyStateMachine6C968CA5-Ybusf26S5Oir"
  }
}

```

This is one of the ways to pass deployed values into the tests. Alternatively, you can also use Zion's `get_stack_outputs` method to retrieve stack outputs.

To clean up the stacks after running the examples:

```bash
npm run destroy
```

### Example02 - Testing EventBridge Event Bus with a "Listener"

Test code:
- [Python](./example02-04/tests/python/test_example_02.py)

This example shows how to use a "Listener" to test a Rule on a given Event Bus. The stack under test is called "cdk-example-ebStack". Three methods are used in this example:

- `add_listener` - create a listener on the provided event bus by replicating a provided rule and target transformation
- `remove_listener` - destroy listener(s)
- `wait_until_event_matched` -  wait until a received event matches the provided `condition` function
- `poll_events` - poll events from the listener and return the received events

### Example03 - Testing end-to-end with X-Ray Traces

Test code:
- [Python](./example02-04/tests/python/test_example_03.py)

This example shows how to test end-to-end with X-Ray Traces. The stack under test is called "cdk-example-sfnStack". Two methods are used in this example:

- `retry_get_trace_tree_until` - retry getting trace tree of given trace ID until it matches the provided `condition`
- `get_trace_tree` - get trace tree of given trace ID. This performs the action once only

