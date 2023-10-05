#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { SfnStack } from '../lib/sfn-stack';


const namespace = process.env.NAMESPACE || 'cdk-example';

const app = new cdk.App();
const sfnStack = new SfnStack(app, `${namespace}-sfnStack`);