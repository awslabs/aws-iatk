# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import os
from distutils.errors import CompileError
from subprocess import call
from typing import List

from setuptools import setup
from setuptools.command.build import build
from setuptools.command.editable_wheel import editable_wheel
from setuptools.command.sdist import sdist

def build_and_install_zion(packages: List[str]) -> None:
    cmd = ['go', 'build', '-C', './src/zion_src', '-o', '../zion_service/', './cmd/zion']
    # TODO (hawflau): introduce env var to control whether to build or not
    if not os.getenv("GOARCH") or not os.path.isfile("./src/zion_service/zion"):
        out = call(cmd)
        if out != 0:
            raise CompileError("Failed to build Zion Service. Golang version >1.20 required and on PATH")

    # Add zion_service package to the packages list. This ensures it is included in the python whl/sdist
    packages.extend(["zion_service"])
    list_to_remove = []
    for package in packages:
        if package.startswith("zion_src"):
            list_to_remove.append(package)

    for to_remove in list_to_remove:
        packages.remove(to_remove)
    
class Build(build):
    def run(self) -> None:
        build_and_install_zion(self.distribution.packages)
        super().run()

class EditableWheel(editable_wheel):
    def run(self) -> None:
        build_and_install_zion(self.distribution.packages)
        super().run()

class Sdist(sdist):
    def run(self) -> None:
        try:
            self.distribution.packages.remove("zion_service")
        except ValueError:
            print("zion_service was not found, hence didn't remove. Continuing with sdist build.")
        super().run()

setup(cmdclass={"build": Build, "editable_wheel": EditableWheel}, 
      # We set this since we are embedding a go binary into the python package.
      # This ensures the whls are platform specific, as the go binary is.
      has_ext_modules=lambda: True)
