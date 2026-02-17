import React from 'react';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

  const handleUpdateProfile = () => {
    navigate('/profile');
  };

  const handleChangePassword = () => {
    // Navigate to profile page where password change can be handled
    navigate('/profile');
  };

  const handleGoToAdmin = () => {
    navigate('/admin');
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="py-8">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h1 className="text-2xl font-bold text-gray-900 mb-6">
              Welcome to your Dashboard
            </h1>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {/* User Info Card */}
              <div className="bg-blue-50 p-6 rounded-lg">
                <h2 className="text-lg font-semibold text-blue-900 mb-4">
                  Your Profile
                </h2>
                <div className="space-y-2">
                  <p className="text-blue-700">
                    <span className="font-medium">Email:</span> {user?.email}
                  </p>
                  <p className="text-blue-700">
                    <span className="font-medium">Username:</span> {user?.username}
                  </p>
                  <p className="text-blue-700">
                    <span className="font-medium">Role:</span> {user?.role}
                  </p>
                </div>
              </div>

              {/* Quick Actions Card */}
              <div className="bg-green-50 p-6 rounded-lg">
                <h2 className="text-lg font-semibold text-green-900 mb-4">
                  Quick Actions
                </h2>
                <div className="space-y-3">
                  <button 
                    onClick={handleUpdateProfile}
                    className="w-full bg-green-600 hover:bg-green-700 text-white py-2 px-4 rounded-md text-sm font-medium transition-colors"
                  >
                    Update Profile
                  </button>
                  <button 
                    onClick={handleChangePassword}
                    className="w-full bg-green-600 hover:bg-green-700 text-white py-2 px-4 rounded-md text-sm font-medium transition-colors"
                  >
                    Change Password
                  </button>
                </div>
              </div>

              {/* Stats Card */}
              <div className="bg-purple-50 p-6 rounded-lg">
                <h2 className="text-lg font-semibold text-purple-900 mb-4">
                  System Status
                </h2>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-purple-700">API Gateway:</span>
                    <span className="text-green-600 font-medium">✓ Online</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-purple-700">Auth Service:</span>
                    <span className="text-green-600 font-medium">✓ Online</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-purple-700">User Service:</span>
                    <span className="text-green-600 font-medium">✓ Online</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Admin Section */}
            {user?.role === 'admin' && (
              <div className="mt-8">
                <div className="bg-yellow-50 p-6 rounded-lg">
                  <h2 className="text-lg font-semibold text-yellow-900 mb-4">
                    Admin Panel
                  </h2>
                  <p className="text-yellow-700 mb-4">
                    You have administrator privileges. Access admin features below.
                  </p>
                  <button 
                    onClick={handleGoToAdmin}
                    className="bg-yellow-600 hover:bg-yellow-700 text-white py-2 px-4 rounded-md text-sm font-medium transition-colors"
                  >
                    Go to Admin Panel
                  </button>
                </div>
              </div>
            )}

            {/* Recent Activity */}
            <div className="mt-8">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Recent Activity
              </h2>
              <div className="bg-gray-50 p-4 rounded-lg">
                <p className="text-gray-600">
                  Welcome to your microservices dashboard! This is where you'll see your recent activity and system updates.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;