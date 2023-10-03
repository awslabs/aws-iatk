#!/usr/bin/bash
archs=(amd64 arm64 ppc64le ppc64 s390x)
pip install build
for arch in ${archs[@]}
do  
    echo ${arch}
	export GOARCH=${arch}
    python3 -m build -w
done