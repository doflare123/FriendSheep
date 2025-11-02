export interface RegisterRequest {
  email: string;
}

export interface RegisterResponse {
  session_id: string;
}

export interface VerifySessionRequest {
  code: string;
  session_id: string;
  type: string;
}

export interface VerifySessionResponse {
  [key: string]: boolean;
}

export interface CreateUserRequest {
  email: string;
  name: string;
  password: string;
  session_id: string;
}

export interface CreateUserResponse {
  [key: string]: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AdminGroup {
  category: string[];
  id: number;
  image: string;
  member_count: number;
  name: string;
  small_description: string;
  type: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  admin_groups: AdminGroup[];
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface ErrorResponse {
  [key: string]: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  tokens: AuthTokens | null;
  adminGroups: AdminGroup[];
}