import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as eventschemas from 'aws-cdk-lib/aws-eventschemas';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as path from 'path';

export class SchemaStack extends cdk.Stack {
    registry: eventschemas.CfnRegistry | null = null;
    schema: eventschemas.CfnSchema | null = null;
    lambdaFunction: lambda.Function | null = null;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        this.registry = new eventschemas.CfnRegistry(this, 'MyRegistry', {});
        this.schema = new eventschemas.CfnSchema(this, 'MySchema', {
            registryName: this.registry.attrRegistryName,
            type: 'OpenApi3',
            content: JSON.stringify({
                openapi: '3.0.0',
                info: {
                    version: '1.0.0',
                    title: 'my-event',
                },
                paths: {},
                components: {
                    schemas: {
                        MyEvent: {
                            type: 'object',
                            properties: {
                                customerId: {
                                    type: 'string',
                                },
                                datetime: {
                                    type: 'string',
                                    format: 'date-time',
                                },
                                membershipType: {
                                    type: 'string',
                                    enum: ['A', 'B', 'C'],
                                },
                                address: {
                                    type: 'string',
                                },
                                orderItems: {
                                    type: 'array',
                                    items: {
                                        $ref: '#/components/schemas/Item',
                                    },
                                },
                            },
                        },
                        Item: {
                            type: 'object',
                            properties: {
                                sku: {
                                    type: 'string',
                                },
                                unitPrice: {
                                    type: 'number',
                                },
                                count: {
                                    type: 'integer',
                                },
                            },
                        },
                    },
                },
            }),
        });

        this.lambdaFunction = new lambda.Function(this, 'Calculator', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'calculatorHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
        });

        // outputs
        new cdk.CfnOutput(this, 'CalculatorFunction', {
            description: 'Lambda Function Name',
            value: this.lambdaFunction.functionName,
        });
        new cdk.CfnOutput(this, 'RegistryName', {
            value: this.registry.attrRegistryName,
        });
        new cdk.CfnOutput(this, 'SchemaName', {
            value: this.schema.attrSchemaName,
        });
    }
}
