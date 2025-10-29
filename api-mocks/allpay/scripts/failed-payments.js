logger.info('running failed payments script')
const scheduleStore = stores.open('schedules');

const schedule = scheduleStore.loadAll();
logger.info(schedule);

if (schedule) {
    const date = schedule.date.replace(/^(\d{4})-(\d{2})-(\d{2})$/, '$3/$2/$1 00:00:00');
    const data = JSON.stringify(
        {
            "FailedPayments": [
                {
                    "Amount": schedule.amount,
                    "ClientReference": base64Decode(schedule.clientRef, 'base64'),
                    "CollectionDate": date,
                    "IsRepresented": true,
                    "LastName": "Jones",
                    "Line1": "Flat 2",
                    "ProcessedDate": date,
                    "ReasonCode": "BACS 0 : Refer to payer",
                    "SchemeCode": "OPGB"
                }
            ],
            "TotalRecords": 1
        }
    )
    respond().withStatusCode(200).withContent(data);
} else {
    respond().withStatusCode(400);
}


function base64Decode(str) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
    const lookup = chars.split('').reduce((acc, c, i) => (acc[c] = i, acc), {});
    let buffer = 0, bits = 0, output = '';

    for (let i = 0; i < str.length; i++) {
        const c = str[i];
        if (c === '=') break;
        const value = lookup[c];
        if (value === undefined) continue;

        buffer = (buffer << 6) | value;
        bits += 6;

        if (bits >= 8) {
            bits -= 8;
            const byte = (buffer >> bits) & 0xFF;
            output += String.fromCharCode(byte);
        }
    }

    return decodeURIComponent(escape(output)); // Converts UTF-8 bytes to string
}
