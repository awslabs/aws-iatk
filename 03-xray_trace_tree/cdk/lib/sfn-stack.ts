import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as sfn from 'aws-cdk-lib/aws-stepfunctions';
import * as tasks from 'aws-cdk-lib/aws-stepfunctions-tasks';
import * as sns from 'aws-cdk-lib/aws-sns';

export class SfnStack extends cdk.Stack {
    statemachine: sfn.StateMachine | null = null;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

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
        this.statemachine = new sfn.StateMachine(this, 'MyStateMachine', {
            definitionBody: sfn.DefinitionBody.fromChainable(definition),
            tracingEnabled: true,
        });

        new cdk.CfnOutput(this, 'StateMachineArn', {
            description: 'State Machine ARN',
            value: this.statemachine.stateMachineArn,
        });
        new cdk.CfnOutput(this, 'StateMachineName', {
            description: 'State Machine Name',
            value: this.statemachine.stateMachineName,
        });
    }
}
