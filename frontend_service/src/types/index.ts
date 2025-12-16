// User types based on the actual API response
export interface User {
  id: string;
  email: string;
  username: string;  // This maps to 'name' from backend
  name: string;      // Backend 'name' field
  role: string;
  created_at: string;
  updated_at: string;
}

// Authentication request/response types
export interface RegisterRequest {
  email: string;
  username: string;  // This will be sent as 'name' to backend
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

// API response matches the actual backend response format
export interface AuthResponse {
  token: string;
  refresh_token: string;  // API uses snake_case
  message?: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;  // API expects snake_case
}

// Admin types
export interface UpdateUserRequest {
  email?: string;
  username?: string;  // This maps to 'name' field in backend
  name?: string;      // Backend 'name' field
  role?: string;
}

export interface DeleteUserRequest {
  userId: string;
}

export interface GetUsersResponse {
  users: User[];
  total: number;
  page: number;
  limit: number;
}

// API Error type
export interface ApiError {
  message: string;
  code: number;
  details?: string;
}

// Auth context types
export interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (email: string, username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<boolean>;
  updateUser: (userData: Partial<User>) => void;
}