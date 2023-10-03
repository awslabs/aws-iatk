#!/usr/bin/bash
pip install build
export GOARCH=${GO_ARCH}
python3 -m build -w