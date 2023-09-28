import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
// import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as events from 'aws-cdk-lib/aws-events';
import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as sfn from 'aws-cdk-lib/aws-stepfunctions';
import * as tasks from 'aws-cdk-lib/aws-stepfunctions-tasks';
import * as sns from 'aws-cdk-lib/aws-sns';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as path from 'path';

export class CdkStack extends cdk.Stack {
    resources: Map<string, string> = new Map();

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        // The code that defines your stack goes here
        const eb = new events.EventBus(this, 'EB', {
            eventBusName: 'my-demo-eb',
        });
        this.resources.set(eb.toString(), this.getLogicalId(eb.node.findChild('Resource') as cdk.CfnElement));

        const producer = new lambda.Function(this, 'Producer', {
            code: lambda.Code.fromAsset(path.resolve('..', 'dist', 'producerHandler')),
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.lambdaHandler',
            environment: {
                EVENTBUS_NAME: eb.eventBusName,
            },
            tracing: lambda.Tracing.ACTIVE,
        });
        eb.grantPutEventsTo(producer);

        const api = new apigateway.RestApi(this, 'myapi', {
            deploy: true,
            deployOptions: {
                tracingEnabled: true,
            },
        });
        const resource = api.root.addResource('orders');
        const integration = new apigateway.LambdaIntegration(producer, {
            proxy: true,
        });
        resource.addMethod('POST', integration);
        resource.addMethod('PUT', integration);

        const sm = this.stateMachine();
        const sfnRole = new iam.Role(this, 'Role', {
            assumedBy: new iam.ServicePrincipal('events.amazonaws.com'),
        });
        const rule = new events.Rule(this, 'ConsumerRule', {
            eventBus: eb,
            eventPattern: {
                source: ['com.hello-world.producer'],
                detailType: ['NewOrder'],
            },
        });
        rule.addTarget(
            new targets.SfnStateMachine(sm, {
                input: events.RuleTargetInput.fromObject({
                    waitMilliseconds: 2000,
                }),
                role: sfnRole,
            }),
        );
    }

    stateMachine(): sfn.StateMachine {
        const convertToSeconds = new tasks.EvaluateExpression(this, 'Convert to seconds', {
            expression: '$.waitMilliseconds / 1000',
            resultPath: '$.waitSeconds',
        });

        const createMessage = new tasks.EvaluateExpression(this, 'Create message', {
            // Note: this is a string inside a string.
            expression: '`Now waiting ${$.waitSeconds} seconds...`',
            runtime: lambda.Runtime.NODEJS_LATEST,
            resultPath: '$.message',
        });

        const publishMessage = new tasks.SnsPublish(this, 'Publish message', {
            topic: new sns.Topic(this, 'cool-topic'),
            message: sfn.TaskInput.fromJsonPathAt('$.message'),
            resultPath: '$.sns',
        });

        const wait = new sfn.Wait(this, 'Wait', {
            time: sfn.WaitTime.secondsPath('$.waitSeconds'),
        });

        const definition = convertToSeconds.next(createMessage).next(publishMessage).next(wait);
        const sm = new sfn.StateMachine(this, 'MyStateMachine', {
            definitionBody: sfn.DefinitionBody.fromChainable(definition),
            tracingEnabled: true,
        });
        return sm;
    }
}
