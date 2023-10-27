---
title: Homepage
description: AWS Cloud Test Kit
---

AWS Cloud Test Kit (CTK) is a framework that makes it easy for developers to write integration tests that run against their Event Driven Application in the cloud. CTK simplifies writing integration tests for serverless applications, providing utilities to generate test events to trigger an application, validate event flow and structure in EventBridge, and assert event flow against X-Ray traces. 

## Install

You can install AWS CTK for Python using one of the following options:

Pip:
```bash
pip install aws-ctk
```

!!! question "Looking for Pip signed releases? [Learn more about verifying signed builds](./security.md#verifying-signed-builds)"

## Quick getting started

```bash title="Hello world example using SAM CLI"
sam init --app-template hello-world-ctk --name sam-app --package-type Zip --runtime python3.11 --no-tracing
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
