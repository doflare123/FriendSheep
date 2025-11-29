import apiClient from '../apiClient';
import { clearTokens, saveTokens } from '../storage/tokenStorage';
import {
  ErrorResponse,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  RegisterRequest,
  RegisterResponse,
  VerifySessionResponse
} from '../types/auth';

class AuthService {
  async createRegistrationSession(
    email: string
  ): Promise<RegisterResponse> {
    try {
      const response = await apiClient.post<RegisterResponse>(
        '/sessions/register',
        { email, timeout: 5000 } as RegisterRequest
      );
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async verifySession(
    sessionId: string,
    code: string,
    type: string = 'register'
  ): Promise<VerifySessionResponse> {
    try {
      const response = await apiClient.patch<VerifySessionResponse>(
        '/sessions/verify',
        {
          session_id: sessionId,
          code: code,
          type: type,
        }
      );
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async createUser(
    email: string,
    name: string,
    password: string,
    sessionId: string
  ): Promise<any> {
    try {
      const response = await apiClient.post('/users', {
        email,
        name,
        password,
        session_id: sessionId,
      });
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async resendCode(email: string): Promise<RegisterResponse> {
    try {
      return await this.createRegistrationSession(email);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async login(email: string, password: string): Promise<LoginResponse> {
    try {
      const response = await apiClient.post<LoginResponse>('/users/login', {
        timeout: 5000,
        email,
        password,
      });

      await saveTokens(
        response.data.access_token,
        response.data.refresh_token
      );

      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }


  async refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    try {
      const response = await apiClient.post<RefreshTokenResponse>(
        '/users/refresh',
        { refresh_token: refreshToken } as RefreshTokenRequest
      );

      await saveTokens(
        response.data.access_token,
        response.data.refresh_token
      );

      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async logout(): Promise<void> {
    try {
      await clearTokens();
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  private handleError(error: any): Error {
    if (error.response) {
      const status = error.response.status;
      const data: ErrorResponse = error.response.data;

      const errorMessage =
        Object.values(data).join(', ') || 'Произошла ошибка';

      switch (status) {
        case 400:
          return new Error(`Неверные данные: ${errorMessage}`);
        case 401:
          return new Error('Неверный код подтверждения или пароль');
        case 404:
          return new Error('Сессия не найдена или истекла');
        case 429:
          return new Error('Слишком много попыток. Попробуйте позже');
        case 500:
          return new Error('Ошибка сервера. Попробуйте позже');
        default:
          return new Error(errorMessage);
      }
    } else if (error.request) {
      return new Error('Нет связи с сервером');
    } else {
      return new Error(error.message || 'Неизвестная ошибка');
    }
  }
}

export default new AuthService();