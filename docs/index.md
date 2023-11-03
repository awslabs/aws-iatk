---
title: AWS Cloud Test Kit Overview
description: AWS Cloud Test Kit
---

!!! alert "AWS Cloud Test Kit is in Public Preview"

AWS Cloud Test Kit (CTK) is a framework that makes it easy for developers to write integration tests that run against their Event Driven Application in the cloud. CTK simplifies writing integration tests for serverless applications, providing utilities to generate test events to trigger an application, validate event flow and structure in EventBridge, and assert event flow against X-Ray traces. 

## Install

You can install AWS CTK for Python using one of the following options:

Pip:
```bash
pip install aws-ctk
```

!!! question "Looking for Pip signed releases? [Learn more about verifying signed builds](./security.md#verifying-signed-builds)"

## Quick getting started

See [Tutorial's](./tutorial/index.md) page for more information.

## Credentials Configurations

AWS CTK requires AWS Credentials in order to interact with AWS Resources in your account. You can specify this in `AWSCtk` directly, as shown below. You can also set [Environment Variables](#environment-variables) instead.

```
from aws_ctk import AWSCtk

ctk = AWSCtk(
	profile=PROFILE,
	region=REGION
)
```

## Environment variables

???+ info
	Explicit parameters take precedence over environment variables

| Enviromment Variable  | Description |
| --------------------- | ----------- |
| AWS_REGION            | AWS Region to use|
| AWS_ACCESS_KEY_ID     | AWS Access Key to use |
| AWS_SECRET_ACCESS_KEY | AWS Secret Access Key to use |
| AWS_SESSION_TOKEN     | AWS Session Token to use (optional) |



## Concepts

### System Under Test (SUT)

The system being tested for correct operations (including happy and error paths).

### Test Harness

A group of AWS resources Testing SDK creates for the purpose of facilitating testing around an integration. These resources are intended to exist only for the duration of the test run, and should be destroyed after the test run completes.

### Arrange, Assert, Act Testing Pattern

AWS CTK enables testing done through the Arrange, Assert, Act Testing Pattern. AWS CTK
will help setup and get deployed resources (Arrange) and give you the information from
in order to Assert on those resources.
