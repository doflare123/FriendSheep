import apiClient from '../apiClient';
import {
    Subscription,
    TileSettings,
    UpdateProfileRequest,
    UpdateProfileResponse,
    UserProfile,
} from '../types/user';

class UserService {
  async getCurrentUserProfile(): Promise<UserProfile> {
    try {
      const response = await apiClient.get<UserProfile>('/users/inf');
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getUserProfileById(userId: string): Promise<UserProfile> {
    try {
      const response = await apiClient.get<UserProfile>(`/users/inf/${userId}`);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getUserSubscriptions(userId?: number): Promise<Subscription[]> {
    try {
      const params = userId ? { id: userId } : {};
      const response = await apiClient.get<Subscription[]>('/users/subscriptions', {
        params,
      });
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateProfile(data: UpdateProfileRequest): Promise<UpdateProfileResponse> {
    try {
      const response = await apiClient.patch<UpdateProfileResponse>(
        '/users/user/profile',
        data
      );
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateTileSettings(settings: TileSettings): Promise<void> {
    try {
      await apiClient.patch('/users/tiles', settings);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  private handleError(error: any): Error {
    if (error.response) {
      const status = error.response.status;
      const data = error.response.data;

      const errorMessage =
        typeof data === 'object'
          ? Object.values(data).join(', ')
          : data || 'Произошла ошибка';

      switch (status) {
        case 400:
          return new Error(`Неверные данные: ${errorMessage}`);
        case 401:
          return new Error('Необходимо войти в систему');
        case 403:
          return new Error('Нет прав для выполнения действия');
        case 404:
          return new Error('Пользователь не найден');
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

export default new UserService();