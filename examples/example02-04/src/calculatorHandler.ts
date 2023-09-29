export const lambdaHandler = async (event, context): Promise<Object> => {
    console.log(event);
    const items = event['orderItems'];
    let total = 0;
    items.forEach((item) => {
        total += item['unitPrice'] * item['count'];
    });
    return total;
};
