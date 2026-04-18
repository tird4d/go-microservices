import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const duration = new Trend('duration');
const successRate = new Rate('success');
const rps = new Counter('rps');

export const options = {
  stages: [
    { duration: '30s', target: 50 },   // ramp up
    { duration: '1m',  target: 50 },   // sustained
    { duration: '10s', target: 150 },  // spike
    { duration: '30s', target: 150 },  // hold spike
    { duration: '30s', target: 0 },    // cool down
  ],
  thresholds: {
    'errors':   ['rate<0.05'],
    'duration': ['p(95)<1000'],
    'success':  ['rate>0.95'],
  },
};

const BASE_URL = 'http://a3c631aa53dd24df3810f28db6f72711-1952013117.eu-central-1.elb.amazonaws.com';

const TEST_USER = {
  name:     'k6authtest',
  email:    'k6authtest@example.com',
  password: 'authtest123',
  role:     'user',
};

// Register once, login to get both tokens
export function setup() {
  http.post(`${BASE_URL}/api/v1/register`, JSON.stringify(TEST_USER), {
    headers: { 'Content-Type': 'application/json' },
  });

  const res = http.post(`${BASE_URL}/api/v1/login`, JSON.stringify({
    email:    TEST_USER.email,
    password: TEST_USER.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const body = JSON.parse(res.body);
  if (!body.token) {
    console.error(`Setup login failed: ${res.body}`);
  }
  return { token: body.token, refresh_token: body.refresh_token };
}

export default function (data) {
  const authHeaders = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
  };

  // 1. Login — most CPU-heavy: bcrypt compare + JWT sign
  group('Login', () => {
    const res = http.post(`${BASE_URL}/api/v1/login`, JSON.stringify({
      email:    TEST_USER.email,
      password: TEST_USER.password,
    }), { headers: { 'Content-Type': 'application/json' } });

    check(res, {
      'login status 200': (r) => r.status === 200,
      'login has token':  (r) => JSON.parse(r.body).token !== undefined,
      'login duration < 1000ms': (r) => r.timings.duration < 1000,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
    sleep(0.5);
  });

  // 2. Validate token via /me — JWT verify on every request
  group('Validate Token (GET /me)', () => {
    const res = http.get(`${BASE_URL}/api/v1/me`, authHeaders);

    check(res, {
      'me status 200': (r) => r.status === 200,
      'me duration < 500ms': (r) => r.timings.duration < 500,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
    sleep(0.3);
  });

  // 3. Refresh token
  group('Refresh Token', () => {
    const res = http.post(`${BASE_URL}/api/v1/refresh-token`, JSON.stringify({
      refresh_token: data.refresh_token,
    }), { headers: { 'Content-Type': 'application/json' } });

    check(res, {
      'refresh status 200': (r) => r.status === 200,
      'refresh has token':  (r) => JSON.parse(r.body).access_token !== undefined,
      'refresh duration < 500ms': (r) => r.timings.duration < 500,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
    sleep(0.3);
  });
}
