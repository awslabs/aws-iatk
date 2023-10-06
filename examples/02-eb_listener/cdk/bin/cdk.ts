#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { EbStack } from '../lib/eb-stack';

const namespace = process.env.NAMESPACE || 'cdk-example';

const app = new cdk.App();
const ebStack = new EbStack(app, `${namespace}-ebStack`);
