import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { Order } from '../types';

const statusColors: Record<string, string> = {
  confirmed: 'bg-green-100 text-green-700',
  pending: 'bg-yellow-100 text-yellow-700',
  cancelled: 'bg-red-100 text-red-700',
};

const Orders: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    apiService.listOrders()
      .then(setOrders)
      .catch(() => setError('Failed to load orders'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex justify-center p-12 text-gray-500">Loading orders…</div>;

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-gray-900">My Orders</h1>
        <button
          onClick={() => navigate('/products')}
          className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 transition"
        >
          Continue Shopping
        </button>
      </div>

      {error && <div className="mb-4 p-3 bg-red-100 text-red-700 rounded">{error}</div>}

      {orders.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="text-lg">No orders yet.</p>
          <button onClick={() => navigate('/products')}
            className="mt-4 text-indigo-600 hover:underline">Browse products →</button>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map(order => (
            <div key={order.id} className="bg-white rounded-lg shadow p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <p className="text-sm text-gray-400">Order ID</p>
                  <p className="font-mono text-sm">{order.id}</p>
                </div>
                <span className={`px-3 py-1 rounded-full text-sm font-medium ${statusColors[order.status] ?? 'bg-gray-100 text-gray-600'}`}>
                  {order.status}
                </span>
              </div>

              <ul className="divide-y mb-4">
                {order.items.map((item, i) => (
                  <li key={i} className="py-2 flex justify-between text-sm">
                    <span>{item.name} <span className="text-gray-400">× {item.quantity}</span></span>
                    <span className="font-medium">€{(item.price * item.quantity).toFixed(2)}</span>
                  </li>
                ))}
              </ul>

              <div className="flex justify-between items-center border-t pt-3">
                <span className="text-sm text-gray-500">
                  {new Date(order.created_at).toLocaleDateString('en-GB', {
                    day: 'numeric', month: 'short', year: 'numeric',
                    hour: '2-digit', minute: '2-digit'
                  })}
                </span>
                <span className="text-lg font-bold text-indigo-600">
                  Total: €{order.total_price.toFixed(2)}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Orders;
