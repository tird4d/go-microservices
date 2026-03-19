import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('product_errors');
const duration = new Trend('product_duration');
const successRate = new Rate('product_success');
const rps = new Counter('product_rps');

export const options = {
  stages: [
    { duration: '30s', target: 50, name: 'ramp-up' },
    { duration: '1m', target: 50, name: 'sustained' },
    { duration: '10s', target: 100, name: 'spike' },
    { duration: '30s', target: 50, name: 'recovery' },
    { duration: '30s', target: 150, name: 'stress' },
    { duration: '30s', target: 0, name: 'cool-down' },
  ],
  thresholds: {
    'product_errors': ['rate<0.1'],
    'product_duration': ['p(95)<400'],
    'product_success': ['rate>0.90'],
  },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
  // List products
  group('List Products', () => {
    let res = http.get(`${BASE_URL}/products`);
    check(res, {
      'status is 200': (r) => r.status === 200,
      'response has products': (r) => r.body.includes('products') || r.body.includes('name'),
      'duration < 300ms': (r) => r.timings.duration < 300,
    });
    duration.add(res.timings.duration);
    successRate.add(res.status === 200);
    errorRate.add(res.status !== 200);
    rps.add(1);
  });

  sleep(0.5);

  // Filter by category
  group('Filter by Category', () => {
    const categories = ['electronics', 'books', 'clothing'];
    const category = categories[Math.floor(Math.random() * categories.length)];
    
    let res = http.get(`${BASE_URL}/products/category/${category}`);
    check(res, {
      'status is 200 or 404': (r) => [200, 404].includes(r.status),
      'duration < 350ms': (r) => r.timings.duration < 350,
    });
    duration.add(res.timings.duration);
    successRate.add([200, 404].includes(res.status));
    errorRate.add(![200, 404].includes(res.status));
    rps.add(1);
  });

  sleep(0.5);

  // Get single product
  group('Get Single Product', () => {
    const productIds = [
      '650f3b8d7f1234567890abcd',
      '650f3b8d7f1234567890abce',
      '650f3b8d7f1234567890abcf',
    ];
    const productId = productIds[Math.floor(Math.random() * productIds.length)];
    
    let res = http.get(`${BASE_URL}/products/${productId}`);
    check(res, {
      'status is 200 or 404': (r) => [200, 404].includes(r.status),
      'duration < 200ms': (r) => r.timings.duration < 200,
    });
    duration.add(res.timings.duration);
    successRate.add([200, 404].includes(res.status));
    errorRate.add(![200, 404].includes(res.status));
    rps.add(1);
  });

  sleep(1);
}
