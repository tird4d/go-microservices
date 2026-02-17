import axios, { AxiosResponse } from 'axios';
import {
  User,
  RegisterRequest,
  LoginRequest,
  AuthResponse,
  RefreshTokenRequest,
  UpdateUserRequest,
  GetUsersResponse,
  ApiError
} from '../types';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Authentication token management
let authToken: string | null = null;

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config) => {
    if (authToken) {
      config.headers.Authorization = `Bearer ${authToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      // Token expired, let the AuthContext handle the refresh
      console.log('Token expired, authentication context will handle refresh');
    }

    return Promise.reject(error);
  }
);

// Main API Service
export const apiService = {
  // Token management
  setAuthToken(token: string | null): void {
    authToken = token;
  },

  // Authentication methods
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response: AxiosResponse<AuthResponse> = await api.post('/login', data);
    return response.data;
  },

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response: AxiosResponse<AuthResponse> = await api.post('/register', data);
    return response.data;
  },

  async logout(): Promise<void> {
    await api.post('/logout');
  },

  async refreshToken(data: RefreshTokenRequest): Promise<AuthResponse> {
    const response: AxiosResponse<AuthResponse> = await api.post('/refresh-token', data);
    return response.data;
  },

  // User methods
  async getCurrentUser(): Promise<User> {
    const response: AxiosResponse<User> = await api.get('/me');
    return response.data;
  },

  async updateProfile(data: UpdateUserRequest): Promise<User> {
    const response: AxiosResponse<User> = await api.put('/me', data);
    return response.data;
  },

  // Admin methods
  async getUsers(page = 1, limit = 10): Promise<GetUsersResponse> {
    const response: AxiosResponse<GetUsersResponse> = await api.get('/admin/users', {
      params: { page, limit }
    });
    return response.data;
  },

  async updateUser(userId: string, data: UpdateUserRequest): Promise<User> {
    const response: AxiosResponse<User> = await api.put(`/admin/users/${userId}`, data);
    return response.data;
  },

  async deleteUser(userId: string): Promise<void> {
    await api.delete(`/admin/users/${userId}`);
  }
};

// Legacy services for backward compatibility
export const authService = {
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return apiService.register(data);
  },

  async login(data: LoginRequest): Promise<AuthResponse> {
    return apiService.login(data);
  },

  async logout(): Promise<void> {
    return apiService.logout();
  },

  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    return apiService.refreshToken({ refresh_token: refreshToken });
  },

  async getMe(): Promise<User> {
    return apiService.getCurrentUser();
  }
};

export const userService = {
  async getProfile(): Promise<User> {
    return apiService.getCurrentUser();
  },

  async updateProfile(data: UpdateUserRequest): Promise<User> {
    return apiService.updateProfile(data);
  }
};

export const adminService = {
  async getUsers(page = 1, limit = 10): Promise<GetUsersResponse> {
    return apiService.getUsers(page, limit);
  },

  async updateUser(userId: string, data: UpdateUserRequest): Promise<User> {
    return apiService.updateUser(userId, data);
  },

  async deleteUser(userId: string): Promise<void> {
    return apiService.deleteUser(userId);
  }
};

// Error handler utility
export const handleApiError = (error: any): ApiError => {
  if (error.response) {
    return {
      message: error.response.data?.message || 'An error occurred',
      code: error.response.status,
      details: error.response.data?.details
    };
  } else if (error.request) {
    return {
      message: 'Network error - please check your connection',
      code: 0
    };
  } else {
    return {
      message: error.message || 'An unexpected error occurred',
      code: -1
    };
  }
};