import { createErrorHandler } from '@/utils/errorHandler';
import { normalizeImageUrl } from '@/utils/imageUtils';
import { sanitizeSearchQuery } from '@/utils/searchSanitizer';
import apiClient from '../apiClient';
import {
  Subscription,
  TileSettings,
  UpdateProfileRequest,
  UpdateProfileResponse,
  UserProfile,
  UserSearchResponse
} from '../types/user';


class UserService {
  private handleError = createErrorHandler('UserService');
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

  async getUserProfileById(userIdOrUsername: string): Promise<UserProfile> {
    try {
      const isNumericId = /^\d+$/.test(userIdOrUsername);
      
      let userId: string;
      
      if (isNumericId) {
        userId = userIdOrUsername;
      } else {
        const usernameResponse = await apiClient.get(`/users/${userIdOrUsername}`);

        if (typeof usernameResponse.data === 'number') {
          userId = usernameResponse.data.toString();
        } else if (usernameResponse.data?.id) {
          userId = usernameResponse.data.id.toString();
        } else {
          throw new Error('Не удалось получить ID пользователя');
        }
      }

      const response = await apiClient.get<UserProfile>(`/users/inf/${userId}`);
      
      const profile = response.data;
      profile.image = normalizeImageUrl(profile.image);

      if (!profile.id) {
        profile.id = parseInt(userId, 10);
      }
      
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

  async searchUsers(query: string, page: number = 1): Promise<UserSearchResponse> {
    try {
      const sanitizedQuery = sanitizeSearchQuery(query);
      
      if (!sanitizedQuery || sanitizedQuery.length < 1) {
        console.log('[UserService] ⚠️ Запрос пустой:', query);
        return {
          users: [],
          total: 0,
          has_more: false,
          page: 1,
        };
      }

      console.log('[UserService] Поиск пользователей по username:', sanitizedQuery);
      
      const response = await apiClient.get<UserSearchResponse>('/users/search', {
        params: { 
          name: sanitizedQuery,
          page: page,
        },
      });

      console.log('[UserService] 📦 Получен ответ:', {
        total: response.data.total,
        users: response.data.users?.length,
      });

      const normalizedUsers = (response.data.users || []).map(user => {
        const originalImage = user.image;
        const normalizedImage = normalizeImageUrl(user.image);
        
        console.log('[UserService] 🖼️ Нормализация изображения:', {
          userId: user.id,
          original: originalImage,
          normalized: normalizedImage,
          changed: originalImage !== normalizedImage,
        });
        
        return {
          ...user,
          image: normalizedImage,
        };
      });

      console.log('[UserService] ✅ Обработано пользователей:', normalizedUsers.length);

      return {
        users: normalizedUsers,
        total: response.data.total || 0,
        has_more: response.data.has_more || false,
        page: response.data.page || page,
      };
    } catch (error: any) {
      console.error('[UserService] ❌ Ошибка поиска пользователей:', error.message);
      throw error;
    }
  }

  async updateProfile(data: UpdateProfileRequest): Promise<UpdateProfileResponse> {
    try {
      console.log('[UserService] 📤 Отправляемые данные профиля:', JSON.stringify(data, null, 2));
      
      const response = await apiClient.patch<UpdateProfileResponse>(
        '/users/user/profile',
        data
      ); 
      
      console.log('[UserService] ✅ Профиль обновлён:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[UserService] ❌ Ошибка обновления профиля:');
      console.error('  Status:', error.response?.status);
      console.error('  Data:', JSON.stringify(error.response?.data, null, 2));
      console.error('  Отправленные данные:', JSON.stringify(data, null, 2));
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