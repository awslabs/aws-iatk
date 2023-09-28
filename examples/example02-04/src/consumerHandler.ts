import { EventBridgeClient, PutEventsCommand } from '@aws-sdk/client-eventbridge'
import { randomUUID } from 'crypto';
import { captureAWSv3Client } from 'aws-xray-sdk';


const EVENTBUS_NAME = process.env.EVENTBUS_NAME;

const ebClient = new EventBridgeClient({ region: process.env.AWS_REGION });
const ebClientPatched = captureAWSv3Client(ebClient as any);

export const newOrderHandler = async (customerId): Promise<Object> => {
    console.log(customerId);
    const orderId = Math.floor(100000 + Math.random() * 900000);
    const input = {
        Entries: [
            {
                Source: 'com.hello-world.new-order-concusmer',
                DetailType: 'CreatedOrder',
                Detail: JSON.stringify({ customerId, orderId }),
                EventBusName: EVENTBUS_NAME,
            }
        ]
    }
    try {
        const command = new PutEventsCommand(input);
        const response = await ebClientPatched.send(command);
        return {
            statusCode: 200,
            body: JSON.stringify({
                message: 'Created new order',
                response: response,
            }),
        };
    } catch (err) {
        console.log(err);
        return {
            statusCode: 500,
            body: JSON.stringify({
                message: `some error happened - ${err}`,
            }),
        };
    }

}