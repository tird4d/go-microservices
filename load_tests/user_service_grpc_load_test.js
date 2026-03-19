import grpc from 'k6/net/grpc';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('grpc_errors');
const duration = new Trend('grpc_duration');
const successRate = new Rate('grpc_success');

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
    'grpc_errors': ['rate<0.1'],
    'grpc_duration': ['p(95)<300'],
    'grpc_success': ['rate>0.90'],
  },
};

const client = new grpc.Client();
client.load(['proto'], 'user.proto');

export default function () {
  const conn = client.connect('localhost:50051', { plaintext: true });

  group('User Service gRPC - GetUser', () => {
    const payload = {
      id: 'user123',
    };

    const response = client.invoke('user.UserService/GetUser', payload, {});
    
    check(response, {
      'status is OK': (r) => r.status === grpc.StatusOK,
      'response has user': (r) => r.message !== null,
      'duration < 150ms': (r) => r.timings.duration < 150,
    });

    duration.add(response.timings.duration);
    successRate.add(response.status === grpc.StatusOK);
    errorRate.add(response.status !== grpc.StatusOK);
  });

  sleep(0.5);

  group('User Service gRPC - ListUsers', () => {
    const response = client.invoke('user.UserService/ListUsers', {}, {});
    
    check(response, {
      'status is OK': (r) => r.status === grpc.StatusOK,
      'response has users': (r) => r.message.users !== undefined,
      'duration < 200ms': (r) => r.timings.duration < 200,
    });

    duration.add(response.timings.duration);
    successRate.add(response.status === grpc.StatusOK);
    errorRate.add(response.status !== grpc.StatusOK);
  });

  conn.close();
}
