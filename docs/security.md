---
title: Security
description: Security practices and processes for AWS IATK
---

<!-- markdownlint-disable MD041 MD043 -->

## Overview

!!! info "We continuously check and evolve our practices, therefore it is possible some diagrams may be eventually consistent."

<!-- --8<-- "SECURITY.md" -->

### Verifying signed builds

#### Terminology

We use [SLSA](https://slsa.dev/spec/v1.0/about){target="_blank" rel="nofollow"} to ensure our builds are reproducible and to adhere to [supply chain security practices](https://slsa.dev/spec/v1.0/threats-overview).

Within our [releases page](https://github.com/awslabs/aws-iatk/releases), you will notice a new metadata file: `multiple.intoto.jsonl`. It's metadata to describe **where**, **when**, and **how** our build artifacts were produced - or simply, **attestation** in SLSA terminology.

For this to be useful, we need a **verification tool** - [SLSA Verifier](https://github.com/slsa-framework/slsa-verifier). SLSA Verifier decodes attestation to confirm the authenticity, identity, and the steps we took in our release pipeline (_e.g., inputs, git commit/branch, GitHub org/repo, build SHA256, etc._).

#### HOWTO

* Download [SLSA Verifier binary](https://github.com/slsa-framework/slsa-verifier#download-the-binary)
* Download the [latest release artifact from PyPi](https://pypi.org/project/aws-iatk/#files) (either wheel or tar.gz )
* Download `python-client.multiple.intoto.jsonl` attestation from the [latest release](https://github.com/awslabs/aws-iatk/releases/latest) under _Assets_

!!! note "Next steps assume macOS on Apple Silicon as the operating system, and release v0.1.0"

You should have the following files in the current directory:

* **SLSA Verifier tool**: `slsa-verifier-darwin-arm64`
* **IATK Python Client Release artifact**: `aws-iatk-0.1.0.tar.gz`
* **IATK Python Client attestation**: `python-client.multiple.intoto.jsonl`

You can now run SLSA Verifier with the following options:

```bash
./slsa-verifier-darwin-arm64 verify-artifact \
    --provenance-path "python-client.multiple.intoto.jsonl" \
    --source-uri github.com/awslabs/aws-iatk \
    aws-iatk-0.1.0.tar.gz
```
