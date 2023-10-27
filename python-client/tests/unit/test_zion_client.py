import boto3
from unittest import TestCase
from unittest.mock import Mock
import aws_ctk

class TestCTKClient(TestCase):
    def setUp(self):
        region = "us-east-1"
        self.clientLambda = boto3.client("lambda", region_name=region)
        self.clientSfn = boto3.client("stepfunctions", region_name=region)
        self.clientApiGateway = boto3.client("apigateway", region_name=region)
        self.ctkClient = aws_ctk.AWSCtk


    def test_event_registered_lambda(self):
        self.assertIsNone(self.clientLambda.meta.events.__dict__.get("_alias_name_cache").get('before-sign.lambda.*'))
        self.ctkClient.patch_aws_client(self,self.clientLambda)
        self.assertIn('before-sign.lambda.*', self.clientLambda.meta.events.__dict__.get("_alias_name_cache"))

    def test_event_registered_api_gateway(self):
        self.assertIsNone(self.clientApiGateway.meta.events.__dict__.get("_alias_name_cache").get('before-sign.apigateway.*'))
        self.ctkClient.patch_aws_client(self,self.clientApiGateway)
        self.assertIn('before-sign.apigateway.*', self.clientApiGateway.meta.events.__dict__.get("_alias_name_cache"))

    def test_event_registered_step_functions(self):
        self.assertIsNone(self.clientSfn.meta.events.__dict__.get("_alias_name_cache").get('before-sign.stepfunctions.*'))
        self.ctkClient.patch_aws_client(self,self.clientSfn)
        self.assertIn('before-sign.stepfunctions.*', self.clientSfn.meta.events.__dict__.get("_alias_name_cache"))

