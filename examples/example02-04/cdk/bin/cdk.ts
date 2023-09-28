#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { CdkStack } from '../lib/cdk-stack';
import { EbStack } from '../lib/eb-stack';
import { SfnStack } from '../lib/sfn-stack';


const namespace = process.env.NAMESPACE || 'example';

const app = new cdk.App();
const ebStack = new EbStack(app, `${namespace}-stack-01`);
const sfnStack = new SfnStack(app, `${namespace}-stack-02`);
