import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { Product } from '../types';
import { useAuth } from '../context/AuthContext';

interface CartItem {
  product: Product;
  quantity: number;
}

const Products: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [cart, setCart] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [ordering, setOrdering] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    apiService.listProducts()
      .then(setProducts)
      .catch(() => setError('Failed to load products'))
      .finally(() => setLoading(false));
  }, []);

  const addToCart = (product: Product) => {
    setCart(prev => {
      const existing = prev.find(i => i.product.id === product.id);
      if (existing) {
        return prev.map(i => i.product.id === product.id
          ? { ...i, quantity: i.quantity + 1 }
          : i);
      }
      return [...prev, { product, quantity: 1 }];
    });
  };

  const removeFromCart = (productId: string) => {
    setCart(prev => prev.filter(i => i.product.id !== productId));
  };

  const cartTotal = cart.reduce((sum, i) => sum + i.product.price * i.quantity, 0);

  const placeOrder = async () => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    setOrdering(true);
    setError('');
    try {
      const order = await apiService.createOrder({
        items: cart.map(i => ({ product_id: i.product.id, quantity: i.quantity })),
      });
      setSuccess(`Order #${order.id.slice(0, 8)}… confirmed! Total: €${order.total_price.toFixed(2)}`);
      setCart([]);
    } catch {
      setError('Failed to place order. Please try again.');
    } finally {
      setOrdering(false);
    }
  };

  if (loading) return <div className="flex justify-center p-12 text-gray-500">Loading products…</div>;

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Store</h1>

      {error && <div className="mb-4 p-3 bg-red-100 text-red-700 rounded">{error}</div>}
      {success && <div className="mb-4 p-3 bg-green-100 text-green-700 rounded">{success}</div>}

      <div className="flex gap-8">
        {/* Product grid */}
        <div className="flex-1">
          {products.length === 0 ? (
            <p className="text-gray-500">No products available yet.</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {products.map(product => (
                <div key={product.id} className="bg-white rounded-lg shadow overflow-hidden flex flex-col">
                  {product.image_url && (
                    <img src={product.image_url} alt={product.name}
                      className="w-full h-48 object-cover" />
                  )}
                  <div className="p-4 flex flex-col flex-1">
                    <h2 className="text-lg font-semibold text-gray-900">{product.name}</h2>
                    <p className="text-sm text-gray-500 mt-1 flex-1">{product.description}</p>
                    <div className="mt-3 flex items-center justify-between">
                      <span className="text-xl font-bold text-indigo-600">€{product.price.toFixed(2)}</span>
                      <span className="text-xs text-gray-400">{product.stock} in stock</span>
                    </div>
                    <button
                      onClick={() => addToCart(product)}
                      disabled={product.stock === 0}
                      className="mt-3 w-full bg-indigo-600 text-white py-2 rounded hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
                    >
                      {product.stock === 0 ? 'Out of stock' : 'Add to cart'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Cart sidebar */}
        {cart.length > 0 && (
          <div className="w-72 flex-shrink-0">
            <div className="bg-white rounded-lg shadow p-4 sticky top-4">
              <h2 className="text-lg font-semibold mb-4">Cart ({cart.length})</h2>
              <ul className="space-y-3 mb-4">
                {cart.map(item => (
                  <li key={item.product.id} className="flex justify-between text-sm">
                    <div>
                      <p className="font-medium">{item.product.name}</p>
                      <p className="text-gray-400">x{item.quantity}</p>
                    </div>
                    <div className="text-right">
                      <p>€{(item.product.price * item.quantity).toFixed(2)}</p>
                      <button onClick={() => removeFromCart(item.product.id)}
                        className="text-red-400 hover:text-red-600 text-xs">remove</button>
                    </div>
                  </li>
                ))}
              </ul>
              <div className="border-t pt-3 mb-4">
                <div className="flex justify-between font-bold">
                  <span>Total</span>
                  <span>€{cartTotal.toFixed(2)}</span>
                </div>
              </div>
              <button
                onClick={placeOrder}
                disabled={ordering}
                className="w-full bg-green-600 text-white py-2 rounded hover:bg-green-700 disabled:opacity-50 transition"
              >
                {ordering ? 'Placing order…' : isAuthenticated ? 'Place Order' : 'Login to Order'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Products;
