---
title: Tutorial
description: Introduction to AWS IATK
---

This tutorial introduces AWS Integrated Application Test Kit (AWS IATK) by going through four examples. Each of them showcases one feature at a time.

For each example, we will execute the following steps:

1. Deploy System Under Test (SUT) with [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html){target="_blank"} or [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html).
2. Run the example test code with [pytest](https://docs.pytest.org/){target="_blank"}.

## Terminologies

Here are some terminologies we will use throughout the examples:

* System Under Test (SUT) - the system being tested for correct operations (including happy and error paths)
* Test Harness - Test Harness is a group of AWS resources AWS IATK creates for the purpose of facilitating testing around an integration. These resources are intended to exist only for the duration of the test run, and should be destroyed after the test run completes.

## Requirements

* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html){target="_blank"} and [configured with your credentials](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-getting-started-set-up-credentials.html){target="_blank"}.
* [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html){target="_blank"} installed.
* [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html) installed.
* [Python 3.8+](https://www.python.org/downloads/) installed.

## Getting started

Clone the examples:

```bash
git clone --single-branch --branch examples git@github.com:awslabs/aws-iatk.git iatk-examples
cd iatk-examples
```

To run the Python (3.8+) examples:

```bash
python -m venv. venv
source .venv/bin/activate

pip install -r requirements.txt
```

