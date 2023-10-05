import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as events from 'aws-cdk-lib/aws-events';
import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as path from 'path';

export class EbStack extends cdk.Stack {
    eventbus: events.EventBus | null = null;
    rule: events.Rule | null = null;
    target: targets.LambdaFunction | null = null;
    api: apigateway.RestApi | null = null;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        this.eventbus = new events.EventBus(this, 'EB');

        const producer = new lambda.Function(this, 'Producer', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'producerHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
            environment: {
                EVENTBUS_NAME: this.eventbus.eventBusName,
            },
            tracing: lambda.Tracing.ACTIVE,
        });
        this.eventbus.grantPutEventsTo(producer);

        this.api = new apigateway.RestApi(this, 'API', {
            deploy: true,
            deployOptions: {
                tracingEnabled: true,
            },
        });
        const resource = this.api.root.addResource('orders');
        const integration = new apigateway.LambdaIntegration(producer, {
            proxy: true,
        });
        resource.addMethod('POST', integration);

        const consumer = new lambda.Function(this, 'Consumer', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'consumerHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
            environment: {
                EVENTBUS_NAME: this.eventbus.eventBusName,
            },
            tracing: lambda.Tracing.ACTIVE,
        });

        this.rule = new events.Rule(this, 'ConsumerRule', {
            eventBus: this.eventbus,
            eventPattern: {
                source: ['com.hello-world.producer'],
                detailType: ['NewOrder'],
            },
        });
        this.target = new targets.LambdaFunction(consumer, {
            event: events.RuleTargetInput.fromEventPath('$.detail.customerId'),
        });
        this.rule.addTarget(this.target);

        this.output();
    }

    output() {
        if (this.eventbus) {
            new cdk.CfnOutput(this, 'EventBusName', {
                description: 'Event Bus Name',
                value: this.eventbus.eventBusName,
            });
        }

        if (this.rule) {
            new cdk.CfnOutput(this, 'RuleName', {
                description: 'Rule Name',
                value: this.rule.ruleName,
            });
            new cdk.CfnOutput(this, 'TargetId', {
                description: 'Target Id',
                value: 'Target0',
            });
        }
        if (this.api) {
            new cdk.CfnOutput(this, 'ApiEndpoint', {
                description: 'API Endpoint',
                value: this.api.urlForPath('/orders'),
            });
        }
    }
}
