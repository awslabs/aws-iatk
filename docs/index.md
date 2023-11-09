---
title: AWS Integrated Application Test Kit Overview
description: AWS Integrated Application Test Kit
---

!!! alert "Integrated Application Test Kit is in Public Preview"

AWS Integrated Application Test Kit (AWS IATK) is a framework that developers can use to write integration tests to run against their event-driven applications in the AWS Cloud. AWS IATK simplifies the writing of integration tests for serverless applications by doing the following:

- Providing utilities that generate test events to trigger an application.
- Validating event flow and structure in Amazon EventBridge.
- Asserting event flow against AWS X-Ray traces. 

## Install

You can install AWS IATK for Python using one of the following options:

Pip:
```bash
pip install aws-iatk
```

!!! question "Looking for Pip signed releases? [Learn more about verifying signed builds](./security.md#verifying-signed-builds)"

## Set up 

### Credentials Configurations

AWS IATK requires AWS credentials in order to interact with the AWS resources in your account. You can specify this in `AwsIatk` directly, as shown below. You can also set [Environment Variables](#environment-variables) instead.

```
from aws_iatk import AwsIatk

iatk = AwsIatk(
	profile=PROFILE,
	region=REGION
)
```

### Environment variables

???+ info
	Explicit parameters take precedence over environment variables

| Enviromment Variable  | Description |
| --------------------- | ----------- |
| AWS_REGION            | AWS Region to use|
| AWS_ACCESS_KEY_ID     | AWS Access Key to use |
| AWS_SECRET_ACCESS_KEY | AWS Secret Access Key to use |
| AWS_SESSION_TOKEN     | AWS Session Token to use (optional) |

## Quick getting started

To start using AWS IATK, see [Tutorial](./tutorial/index.md).

## Concepts

### System Under Test (SUT)

The system being tested for correct operations (including happy and error paths).

### Test Harness

A group of AWS resources AWS IATK creates for the purpose of facilitating testing around an integration. These resources are intended to exist only for the duration of the test run, and should be destroyed after the test run completes.

### Arrange, Act, Assert Testing Pattern

AWS IATK enables testing done through the Arrange, Act, Assert testing pattern. AWS IATK will help setup and retrieve deployed resources (Arrange). Then, AWS IATK gives you the information in order to Assert on those resources.