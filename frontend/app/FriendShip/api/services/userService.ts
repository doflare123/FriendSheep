import { normalizeImageUrl } from '@/utils/imageUtils';
import apiClient from '../apiClient';
import {
  SearchUsersResponse, // 🆕
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
      const profile = response.data;

      profile.image = normalizeImageUrl(profile.image);
      
      console.log('[NORMALIZED PROFILE IMAGE]', profile.image);
      
      return profile;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getUserProfileById(userId: string): Promise<UserProfile> {
    try {
      const response = await apiClient.get<UserProfile>(`/users/inf/${userId}`);
      const profile = response.data;

      profile.image = normalizeImageUrl(profile.image);
      
      return profile;
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
      
      const normalizedSubscriptions = response.data.map(sub => ({
        ...sub,
        image: normalizeImageUrl(sub.image),
      }));
      
      console.log('[UserService] Подписки загружены:', normalizedSubscriptions.length);
      
      return normalizedSubscriptions;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async searchUsers(query: string, page: number = 1): Promise<SearchUsersResponse> {
    try {
      console.log('[UserService] 🔍 Поиск пользователей:', { query, page });
      
      const response = await apiClient.get<SearchUsersResponse>('/users/search', {
        params: {
          name: query,
          page: page,
        },
      });

      const users = response.data.users || [];

      const normalizedUsers = users.map(user => ({
        ...user,
        image: normalizeImageUrl(user.image),
      }));

      console.log('[UserService] ✅ Найдено пользователей:', normalizedUsers.length);

      return {
        total: response.data.total || 0,
        page: response.data.page || 1,
        has_more: response.data.has_more || false,
        users: normalizedUsers,
      };
    } catch (error: any) {
      console.error('[UserService] ❌ Ошибка поиска:', error);
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

  async uploadImage(imageUri: string): Promise<string> {
    try {
      const formData = new FormData();
      
      const uriParts = imageUri.split('.');
      const fileExtension = uriParts[uriParts.length - 1].toLowerCase();
      
      formData.append('image', {
        uri: imageUri,
        name: `upload_${Date.now()}.${fileExtension}`,
        type: `image/${fileExtension === 'jpg' ? 'jpeg' : fileExtension}`,
      } as any);

      const response = await apiClient.post('/admin/groups/UploadPhoto', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      console.log('[UserService] Ответ сервера:', response.data);

      const imageUrl = response.data.image || 
                      response.data.url || 
                      response.data.image_url ||
                      response.data.path;

      if (!imageUrl || typeof imageUrl !== 'string') {
        console.error('[UserService] Неожиданный ответ:', response.data);
        throw new Error('Сервер не вернул URL изображения');
      }

      return imageUrl;
    } catch (error: any) {
      console.error('[UserService] Ошибка загрузки изображения:', error.response?.data);
      throw this.handleError(error);
    }
  }
}

export default new UserService();