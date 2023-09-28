# Examples

Below lists the examples to showcase how you can use Zion to write integration against the cloud more easily.

To run the examples in Python (3.7+):
```
$ cd examples

# setup venv
$ python -m venv .venv
$ source .venv/bin/activate

# install dependencies
$ pip install -r requirements.txt
```

## Example01 - retrieving information from a deployed stack

This example shows how to use `get_stack_outputs` and `get_physical_id_from_stack` to retrieve information from a deployed stack. They are useful if you deploy your stack directly with a CloudFormation template.

We will use SAM CLI to deploy a [stack](./examples/example01/template.json) to CloudFormation. Then we will use `pytest` to run the [test code](./examples/example01/test_example_01.py).

To run the example:

```bash
# To deploy the stack under test using SAM CLI:
$ sam deploy --stack-name example-01 --template ./examples/example01/template.json

# After deploy completes, run:
$ pytest examples/example01
```



