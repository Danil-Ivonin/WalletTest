import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        mixed_same_wallet_load: {
            executor: 'constant-arrival-rate',
            rate: Number(__ENV.RATE || 1000),
            timeUnit: '1s',
            duration: __ENV.DURATION || '30s',
            preAllocatedVUs: Number(__ENV.VUS || 300),
            maxVUs: Number(__ENV.MAX_VUS || 1000),
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<5000'],
        'checks{check:no_5xx}': ['rate==1'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';
const WALLET_ID = __ENV.WALLET_ID || '11111111-1111-1111-1111-111111111111';

export function setup() {
    // Pre-fund wallet so WITHDRAW requests mostly succeed.
    const payload = JSON.stringify({
        valletId: WALLET_ID,
        operationType: 'DEPOSIT',
        amount: 1000000,
    });

    const res = http.post(`${BASE_URL}/api/v1/wallet`, payload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(res, {
        'setup deposit accepted': (r) => r.status === 200,
    });

    return { walletId: WALLET_ID };
}

export default function (data) {
    const walletId = data.walletId;
    const roll = Math.random();

    let res;

    if (roll < 0.4) {
        // 40% DEPOSIT
        res = postOperation(walletId, 'DEPOSIT', 1);
    } else if (roll < 0.75) {
        // 35% WITHDRAW
        res = postOperation(walletId, 'WITHDRAW', 1);
    } else {
        // 25% GET balance
        res = http.get(`${BASE_URL}/api/v1/wallets/${walletId}`, {
            tags: { endpoint: 'get_balance' },
        });
    }

    check(
        res,
        {
            'no 5xx': (r) => r.status < 500,
            'expected status': (r) => [200, 400, 404, 409, 408, 429].includes(r.status),
        },
        { check: 'no_5xx' },
    );

    sleep(0.001);
}

function postOperation(walletId, operationType, amount) {
    const payload = JSON.stringify({
        valletId: walletId,
        operationType,
        amount,
    });

    return http.post(`${BASE_URL}/api/v1/wallet`, payload, {
        headers: { 'Content-Type': 'application/json' },
        tags: { endpoint: `post_${operationType.toLowerCase()}` },
    });
}

export function teardown(data) {
    const res = http.get(`${BASE_URL}/api/v1/wallets/${data.walletId}`);
    console.log(`Final balance response: status=${res.status}, body=${res.body}`);
}
