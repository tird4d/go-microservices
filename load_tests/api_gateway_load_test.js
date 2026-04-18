import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const duration = new Trend('duration');
const successRate = new Rate('success');
const rps = new Counter('rps');

// Test configuration
export const options = {
  stages: [
    // Ramp up from 0 to 100 users over 30 seconds
    { duration: '30s', target: 100, name: 'ramp-up' },
    // Stay at 100 users for 1 minute
    { duration: '1m', target: 100, name: 'sustained' },
    // Spike to 200 users (sudden traffic)
    { duration: '10s', target: 200, name: 'spike' },
    // Back to 100 users
    { duration: '30s', target: 100, name: 'recovery' },
    // Stress test - ramp up to 300 users
    { duration: '30s', target: 300, name: 'stress' },
    // Cool down
    { duration: '30s', target: 0, name: 'cool-down' },
  ],
  thresholds: {
    'errors': ['rate<0.05'],
    'duration': ['p(95)<500'],
    'success': ['rate>0.95'],
  },
};

const BASE_URL = 'http://a3c631aa53dd24df3810f28db6f72711-1952013117.eu-central-1.elb.amazonaws.com';

const TEST_USER = {
  name: 'k6loadtest',
  email: 'k6loadtest@example.com',
  password: 'loadtest123',
  role: 'user',
};

// setup() runs once before all VUs start — registers the test user and logs in
export function setup() {
  // Register test user (ignore 500 if already exists)
  http.post(`${BASE_URL}/api/v1/register`, JSON.stringify(TEST_USER), {
    headers: { 'Content-Type': 'application/json' },
  });

  // Login to get token
  const loginRes = http.post(`${BASE_URL}/api/v1/login`, JSON.stringify({
    email: TEST_USER.email,
    password: TEST_USER.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const body = JSON.parse(loginRes.body);
  if (!body.token) {
    console.error(`Login failed: ${loginRes.body}`);
  }
  return { token: body.token };
}

export default function (data) {
  const authHeaders = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
  };

  // Group 1: Product endpoints (most common, public)
  group('Product Operations', () => {
    // List products
    let res = http.get(`${BASE_URL}/api/v1/products`);
    check(res, {
      'list products status is 200': (r) => r.status === 200,
      'list products duration < 200ms': (r) => r.timings.duration < 200,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
    sleep(0.5);

    // Get single product
    res = http.get(`${BASE_URL}/api/v1/products/650f3b8d7f1234567890abcd`);
    check(res, {
      'get product status is 200 or 404': (r) => r.status === 200 || r.status === 404,
      'get product duration < 150ms': (r) => r.timings.duration < 150,
    });
    duration.add(res.timings.duration);
    successRate.add([200, 404].includes(res.status));
    errorRate.add(![200, 404].includes(res.status));
    rps.add(1);
    sleep(0.5);

    // Get products by category
    res = http.get(`${BASE_URL}/api/v1/products/category/electronics`);
    check(res, {
      'filter by category status is 200 or 404': (r) => r.status === 200 || r.status === 404,
      'filter by category duration < 300ms': (r) => r.timings.duration < 300,
    });
    duration.add(res.timings.duration);
    successRate.add([200, 404].includes(res.status));
    errorRate.add(![200, 404].includes(res.status));
    rps.add(1);
    sleep(0.5);
  });

  // Group 2: Authenticated user endpoints
  group('User Operations', () => {
    // GET /api/v1/me — requires valid JWT
    const res = http.get(`${BASE_URL}/api/v1/me`, authHeaders);
    check(res, {
      'get me status is 200': (r) => r.status === 200,
      'get me duration < 300ms': (r) => r.timings.duration < 300,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
    sleep(0.5);
  });

  // Group 3: Health checks
  group('Health Checks', () => {
    const res = http.get(`${BASE_URL}/health`);
    check(res, {
      'health check status is 200': (r) => r.status === 200,
      'health check duration < 50ms': (r) => r.timings.duration < 50,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
    sleep(1);
  });
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    '/tmp/k6-api-gateway-results.json': JSON.stringify(data),
  };
}

// Simple text summary function
function textSummary(data, options) {
  const indent = options.indent || '';
  let summary = '\n✅ API Gateway Load Test Summary\n';
  summary += '================================\n\n';

  if (data.metrics) {
    summary += 'Metrics:\n';
    Object.entries(data.metrics).forEach(([name, metric]) => {
      if (metric.values && Object.keys(metric.values).length > 0) {
        summary += `${indent}${name}: ${JSON.stringify(metric.values).substring(0, 50)}...\n`;
      }
    });
  }

  return summary;
}
