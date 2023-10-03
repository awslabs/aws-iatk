#!/usr/bin/bash
archs=(amd64 arm64 ppc64le ppc64 s390x)

for arch in ${archs[@]}
do  
    pip install build
	export GOARCH=${arch}
    python3 -m build -w
done