import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 10 },   // Ramp-up to 10 users over 30 seconds
        { duration: '1m', target: 10 },    // Stay at 10 users for 1 minute
        { duration: '30s', target: 0 },    // Ramp-down to 0 users over 30 seconds
    ],
};

export default function () {
    const url = 'http://localhost:8080/pay';

    let res = http.get(url);

    // Check if the response status is 200 OK
    check(res, {
        'is status 200': (r) => r.status === 200,
    });

    // Pause between requests for 1 second to simulate real user activity
    sleep(1);
}
