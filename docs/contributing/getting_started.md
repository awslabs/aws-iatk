Great you want to contribute to AWS Cloud Test Kit (AWS CTK)!

Before moving forward, see [Conventions](./conventions.md) to learn about our coding practices, coding style, and package setup.

## Overview

AWS CTK is built as a mono repo, in which we have both a RPC binary (the Golang code) and the "language client" (currently Python). 

Everything under `python-client` is the Python related code. We embed the Go binary into this at release but locally you can build the binary and install the Python Client as needed.

Typically, changes will be focused within the Go code. This holds the code that powers AWS CTK and provides easy language expansion.

## How to contribute

We accept all contributions, such as documentation updates, code refactors, and new features. However, the maintainers (AWS) reserve the right to reject contributions that don't fit our long term vision, are complicated to maintain, or other reasons. We will do our best to communicate why something is not accepted. To reduce churn, we ask that your contributions start from a GitHub issue. Simple tasks, like grammar, fixing language in docs, etc do not require a GitHub issue. If the issue needs alignment with the team (e.g feature or bug), we will close the PR and ask for a GitHub Issue. This is to reduce time and wasted effort fixing or adding something we might not accept. You can find issues to work on by looking at `stage/accepted` labels that are not assigned. This indicates that the issue was reviewed by the team and ready for anyone to pickup (team member or community).

Once you have an issue, make changes as needed. We encourage the community to submit Draft PRs which can allow a maintainer to help in the event you are stuck.

Once changes are ready, either submit the PR or move the open PR out of draft. A maintainer will then review the code and provide targeted feedback. 

Once a maintainer or maintainers approve, the PR will be merged and slotted for a release. We do not communicate releases ahead of time, as there are external factors to a release that cannot be accounted for. We will do our best to ship updates regularly to ensure fixes and features land in customers hands.

