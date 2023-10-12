## Terms

These are some important terms we use in Testing SDK:

•	System Under Test - the system being tested for correct operations (including happy and error paths)

•	Test Harness - Test Harness is a group of AWS resources Testing SDK creates for the purpose of facilitating testing around an integration. These resources are intended to exist only for the duration of the test run, and should be destroyed after the test run completes.

## Overview

Testing SDK is a library used in your test code. See examples below for snippets of using Testing SDK in your Python test code.

For more detailed docs on the Python modules [see](../api/python)

## Gerenal Flow of Tests
Here is a general flow to run a test written with Testing SDK:

1.	Deploy System Under Test with your choice of tool (e.g. SAM CLI, CDK, Terraform, etc)

2.	Run the test

    a.	Test creates Test Harness resources

    b.	Test runs test cases

    c.	Test tears down Test Harness resources

3.	Tear down System Under Test

Here is a recommended pattern assuming you use Python’s [`unittest.TestCase`](https://docs.python.org/3/library/unittest.html#unittest.TestCase):
•	Create Test Harness resources in `setUpClass` method or in `setUp` method
•	Tear down Test Harness resources in [`tearDownClass`](https://docs.python.org/3/library/unittest.html#unittest.TestCase.tearDownClass) method or in [`tearDown`](https://docs.python.org/3/library/unittest.html#unittest.TestCase.tearDown) method