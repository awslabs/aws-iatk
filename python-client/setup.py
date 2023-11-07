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

def build_and_install_ctk(packages: List[str]) -> None:
    cmd = ['go', 'build', '-C', './src/ctk_src', '-o', '../ctk_service/', './cmd/ctk']
    if not os.getenv("CTK_SKIP_BUILD_BINARY"):
        out = call(cmd)
        if out != 0:
            raise CompileError("Failed to build CTK Service. Golang version >1.20 required and on PATH")

    # Add ctk_service package to the packages list. This ensures it is included in the python whl/sdist
    packages.extend(["ctk_service"])
    list_to_remove = []
    for package in packages:
        if package.startswith("ctk_src"):
            list_to_remove.append(package)

    for to_remove in list_to_remove:
        packages.remove(to_remove)
    
class Build(build):
    def run(self) -> None:
        build_and_install_ctk(self.distribution.packages)
        super().run()

class EditableWheel(editable_wheel):
    def run(self) -> None:
        build_and_install_ctk(self.distribution.packages)
        super().run()

class Sdist(sdist):
    def run(self) -> None:
        try:
            self.distribution.packages.remove("ctk_service")
        except ValueError:
            print("ctk_service was not found, hence didn't remove. Continuing with sdist build.")
        super().run()

setup(cmdclass={"build": Build, "editable_wheel": EditableWheel}, 
      # We set this since we are embedding a go binary into the python package.
      # This ensures the whls are platform specific, as the go binary is.
      has_ext_modules=lambda: True)
