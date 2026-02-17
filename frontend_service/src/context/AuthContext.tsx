import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiService } from '../services/api';
import { User } from '../types';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (email: string, username: string, password: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<boolean>;
  updateUser: (userData: Partial<User>) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const isAuthenticated = !!token && !!user;

  // Load token and user from localStorage on app start
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        const storedToken = localStorage.getItem('token');
        const storedRefreshToken = localStorage.getItem('refreshToken');
        
        if (storedToken && storedRefreshToken) {
          setToken(storedToken);
          apiService.setAuthToken(storedToken);
          
          // Try to fetch user data to validate token
          try {
            const userData = await apiService.getCurrentUser();
            setUser(userData);
          } catch (error) {
            // Token might be expired, try to refresh
            console.log('Token validation failed, attempting refresh...');
            const refreshSuccess = await refreshTokenInternal();
            if (!refreshSuccess) {
              // Refresh failed, clear auth data
              clearAuthData();
            }
          }
        }
      } catch (error) {
        console.error('Auth initialization error:', error);
        clearAuthData();
      } finally {
        setIsLoading(false);
      }
    };

    initializeAuth();
  }, []);

  const clearAuthData = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    apiService.setAuthToken(null);
  };

  const login = async (email: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      setIsLoading(true);
      const response = await apiService.login({ email, password });
      
      if (response.token && response.refresh_token) {
        // Store tokens
        localStorage.setItem('token', response.token);
        localStorage.setItem('refreshToken', response.refresh_token);
        
        // Set token in API service
        apiService.setAuthToken(response.token);
        setToken(response.token);
        
        // Fetch user data
        try {
          const userData = await apiService.getCurrentUser();
          setUser(userData);
          return { success: true };
        } catch (userError) {
          console.error('Failed to fetch user data after login:', userError);
          return { success: false, error: 'Failed to load user data' };
        }
      } else {
        return { success: false, error: 'Invalid response from server' };
      }
    } catch (error: any) {
      console.error('Login error:', error);
      return { 
        success: false, 
        error: error.response?.data?.message || error.message || 'Login failed' 
      };
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (email: string, username: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      setIsLoading(true);
      const response = await apiService.register({ email, username, password });
      
      if (response.token && response.refresh_token) {
        // Store tokens
        localStorage.setItem('token', response.token);
        localStorage.setItem('refreshToken', response.refresh_token);
        
        // Set token in API service
        apiService.setAuthToken(response.token);
        setToken(response.token);
        
        // Fetch user data
        try {
          const userData = await apiService.getCurrentUser();
          setUser(userData);
          return { success: true };
        } catch (userError) {
          console.error('Failed to fetch user data after registration:', userError);
          return { success: false, error: 'Failed to load user data' };
        }
      } else {
        return { success: false, error: 'Invalid response from server' };
      }
    } catch (error: any) {
      console.error('Registration error:', error);
      return { 
        success: false, 
        error: error.response?.data?.message || error.message || 'Registration failed' 
      };
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    try {
      // Call logout API if possible
      if (token) {
        try {
          await apiService.logout();
        } catch (error) {
          console.error('Logout API call failed:', error);
        }
      }
    } finally {
      clearAuthData();
    }
  };

  const refreshTokenInternal = async (): Promise<boolean> => {
    try {
      const storedRefreshToken = localStorage.getItem('refreshToken');
      if (!storedRefreshToken) {
        return false;
      }

      const response = await apiService.refreshToken({ refresh_token: storedRefreshToken });
      
      if (response.token && response.refresh_token) {
        localStorage.setItem('token', response.token);
        localStorage.setItem('refreshToken', response.refresh_token);
        
        apiService.setAuthToken(response.token);
        setToken(response.token);
        
        // Fetch updated user data
        const userData = await apiService.getCurrentUser();
        setUser(userData);
        
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('Token refresh failed:', error);
      return false;
    }
  };

  const refreshToken = async (): Promise<boolean> => {
    return refreshTokenInternal();
  };

  const updateUser = (userData: Partial<User>) => {
    if (user) {
      setUser({ ...user, ...userData });
    }
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated,
    isLoading,
    login,
    register,
    logout,
    refreshToken,
    updateUser,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};