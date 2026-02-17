import React from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';

const Home: React.FC = () => {
  const { user, isAuthenticated } = useAuth();

  return (
    <div className="px-4 py-6 sm:px-0">
      <div className="border-4 border-dashed border-gray-200 rounded-lg p-8">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-6">
            Welcome to Microservices App
          </h1>
          
          {isAuthenticated ? (
            <div className="space-y-4">
              <p className="text-xl text-gray-600">
                Hello, {user?.username}! You are successfully logged in.
              </p>
              
              <div className="bg-white shadow rounded-lg p-6 max-w-md mx-auto">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">
                  Your Account Information
                </h2>
                <div className="space-y-2 text-left">
                  <p><span className="font-medium">Email:</span> {user?.email}</p>
                  <p><span className="font-medium">Username:</span> {user?.username}</p>
                  <p><span className="font-medium">Role:</span> {user?.role}</p>
                  <p><span className="font-medium">Member since:</span> {
                    user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'
                  }</p>
                </div>
              </div>
              
              <div className="flex justify-center space-x-4 mt-6">
                <Link
                  to="/dashboard"
                  className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-md text-sm font-medium"
                >
                  Go to Dashboard
                </Link>
                <Link
                  to="/profile"
                  className="bg-gray-600 hover:bg-gray-700 text-white px-6 py-2 rounded-md text-sm font-medium"
                >
                  View Profile
                </Link>
                {user?.role === 'admin' && (
                  <Link
                    to="/admin"
                    className="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded-md text-sm font-medium"
                  >
                    Admin Panel
                  </Link>
                )}
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-xl text-gray-600">
                A modern microservices application built with React and Go.
              </p>
              
              <div className="bg-white shadow rounded-lg p-6 max-w-md mx-auto">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">
                  Features
                </h2>
                <ul className="text-left space-y-2 text-gray-600">
                  <li>• User authentication with JWT tokens</li>
                  <li>• Role-based access control</li>
                  <li>• Admin user management</li>
                  <li>• Secure API endpoints</li>
                  <li>• Responsive design</li>
                </ul>
              </div>
              
              <div className="flex justify-center space-x-4 mt-6">
                <Link
                  to="/login"
                  className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-md text-sm font-medium"
                >
                  Sign In
                </Link>
                <Link
                  to="/register"
                  className="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded-md text-sm font-medium"
                >
                  Create Account
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Home;