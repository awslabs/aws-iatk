import { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';
import { EventBridgeClient, PutEventsCommand } from '@aws-sdk/client-eventbridge'
import { captureAWSv3Client } from 'aws-xray-sdk';

const EVENTBUS_NAME = process.env.EVENTBUS_NAME;

const ebClient = new EventBridgeClient({ region: process.env.AWS_REGION });
const ebClientPatched = captureAWSv3Client(ebClient as any);

/**
 *
 * Event doc: https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format
 * @param {Object} event - API Gateway Lambda Proxy Input Format
 *
 * Return doc: https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html
 * @returns {Object} object - API Gateway Lambda Proxy Output Format
 *
 */

export const lambdaHandler = async (event: APIGatewayProxyEvent): Promise<APIGatewayProxyResult> => {
    try {
        if (event.httpMethod === 'PUT' && event.resource === '/orders/{orderId}') {
            const input = {
                Entries: [
                    {
                        Source: 'com.hello-world.producer',
                        DetailType: 'UpdateOrder',
                        Detail: JSON.stringify({ orderId: event.pathParameters.orderId, status: 'Complete' }),
                        EventBusName: EVENTBUS_NAME,
                    },
                ],
            };
            const command = new PutEventsCommand(input);
            const response = await ebClientPatched.send(command);
            return {
                statusCode: 200,
                body: JSON.stringify({
                    message: 'Requested to update order',
                    response: response,
                    // event: event,
                }),
            };
        } else if (event.httpMethod === 'POST' && event.resource === '/orders') {
            if (!event.queryStringParameters || !event.queryStringParameters.customerId) {
                throw new Error("customerId is required");
            }
            const customerId = event.queryStringParameters.customerId;
            const input = {
                Entries: [
                    {
                        Source: 'com.hello-world.producer',
                        DetailType: 'NewOrder',
                        Detail: JSON.stringify({ customerId }),
                        EventBusName: EVENTBUS_NAME,
                    },
                ],
            };
            const command = new PutEventsCommand(input);
            const response = await ebClientPatched.send(command);
            return {
                statusCode: 200,
                body: JSON.stringify({
                    message: 'Requested to create new order',
                    response: response,
                    event: event,
                }),
            };
        }
        throw new Error('Unknown path and method');
    } catch (err) {
        console.log(err);
        return {
            statusCode: 500,
            body: JSON.stringify({
                message: `some error happened - ${err}`,
            }),
        };
    }
};
