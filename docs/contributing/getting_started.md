Great you want to contribute to AWS Cloud Test Kit!

Please see [conventions](./conventions.md) to learn about coding practices and style we expect and setup to learn about general setup fo the package.

## Overview

AWS Cloud Test Kit (CTK) is built as a mono repo, in which we have both a RPC binary (the Golang code) and the "language client" (currently Python). 

Everything under `python-client` is the Python related code. We embed the Go binary into this at release but locally you can build the binary and install the Python Client as needed.

Typically, changes will be focused within the Go code. This holds what powers AWS CTK and is done to easy language expansion.

## How to contribute

We accept all contributions from doc updates to code refactors to new features. However, the maintainers (AWS) reserve the right to reject contributions that do not fit our long term visions, are complicated to maintain, or anything else. We will do our best to communicate why something is not accepted. To reduce churn, we ask contributions start from a Github Issue. Simple tasks like grammer, fixing language in docs, etc do not require one. If the issue needs alignment with the team (e.g feature or bug), we will close the PR and ask for a Github Issue. This is to reduce time and wasted effort fixing or adding something we might not accept. You can find issue to work on by looking at `stage/accepted` labels that are not assigned. This indicates the issue was reviewed by the team and ready for anyone to pickup (team member or community).

Once you have an issue, make changes as needed. We encourage the community to submit Draft PRs which can allow a maintainer to help in the event you are stuck.

Once changes are ready, either submit the PR or move the open PR out of Draft. A maintainer will then look at review the code and providing targeted feedback. 

Once a maintainer or maintainers approve, the PR will be merged and sloted for a release. We do not communicate releases ahead of time, as there are external factors to a release that cannot be accounted for. We will do our best to ship updates regularly to ensure fixes and features land in customers hands.

