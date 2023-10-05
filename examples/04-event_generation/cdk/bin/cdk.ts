#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { SchemaStack } from '../lib/schema-stack';


const namespace = process.env.NAMESPACE || 'cdk-example';

const app = new cdk.App();
const schemaStack = new SchemaStack(app, `${namespace}-schemaStack`);
