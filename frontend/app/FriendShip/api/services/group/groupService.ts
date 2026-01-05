import apiClient from '@/api/apiClient';
import { fetchWithRetry } from '@/utils/errorHandler';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import {
  createGroupFormData,
  createPhotoFormData,
  extractImageUrl,
  handleApiError,
  validateCreateGroupData,
} from './groupHelpers';
import {
  AdminGroup,
  CreateGroupData,
  GroupDetailResponse,
  PublicGroupResponse,
  UpdateGroupData,
} from './groupTypes';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

class GroupService {

  async createGroup(data: CreateGroupData): Promise<any> {
    validateCreateGroupData(data);
    console.log('[GroupService] Создание группы:', data.name);

    const formData = createGroupFormData(data);

    try {
      const result = await fetchWithRetry(
        `${BASE_URL}/groups/createGroup`,
        {
          method: 'POST',
          body: formData,
        },
        'GroupService.createGroup'
      );

      console.log('[GroupService] Группа успешно создана');
      return result;
    } catch (error: any) {
      throw error;
    }
  }

  async getAdminGroups(): Promise<AdminGroup[]> {
    try {
      const response = await apiClient.get('/admin/groups');
      console.log('Ответ от API (admin/groups):', response.data);

      if (!response.data || !Array.isArray(response.data)) {
        console.warn('API вернул некорректные данные:', response.data);
        return [];
      }
      
      return response.data;
    } catch (error: any) {
      console.error('Ошибка получения групп администратора:', error);
      console.error('Детали ошибки:', error.response?.data);
      throw new Error(error.response?.data?.message || 'Ошибка получения групп');
    }
  }

  async getGroupDetail(groupId: string | number): Promise<GroupDetailResponse> {
    try {
      const endpoint = `/admin/groups/${groupId}/infGroup`;
      console.log(`Загрузка детальной информации о группе ${groupId}...`);
      console.log(`Полный URL: ${BASE_URL}${endpoint}`);
      
      const response = await apiClient.get(endpoint);
      console.log('Детальная информация о группе:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('Ошибка получения детальной информации о группе:', error);
      console.error('Статус:', error.response?.status);
      console.error('Детали ошибки:', error.response?.data);
      
      throw new Error(error.response?.data?.message || 'Ошибка получения информации о группе');
    }
  }

  async getPublicGroupDetail(groupId: string | number): Promise<PublicGroupResponse> {
    try {
      console.log(`Загрузка публичной информации о группе ${groupId}...`);
      const response = await apiClient.get(`/groups/${groupId}`);
      console.log('Публичная информация о группе:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('Ошибка получения публичной информации о группе:', error);
      console.error('Детали ошибки:', error.response?.data);
      throw error;
    }
  }

  async uploadGroupPhoto(imageUri: string): Promise<string> {
    try {
      const formData = createPhotoFormData(imageUri);
      console.log('Загрузка изображения группы...');

      const result = await fetchWithRetry(
        `${BASE_URL}/admin/groups/UploadPhoto`,
        {
          method: 'POST',
          body: formData,
        },
        'GroupService.uploadGroupPhoto'
      );

      const imageUrl = extractImageUrl(result);
      console.log('URL загруженного изображения:', imageUrl);
      return imageUrl;
    } catch (error: any) {
      throw error;
    }
  }

  async updateGroup(groupId: string | number, data: UpdateGroupData): Promise<any> {
    try {
      console.log('Обновление группы:', groupId, data);
      const response = await apiClient.patch(`/admin/groups/${groupId}`, data);
      console.log('Группа успешно обновлена:', response.data);
      return response.data;
    } catch (error: any) {
      return handleApiError(error, 'Ошибка обновления группы');
    }
  }

  async deleteGroup(groupId: string | number): Promise<void> {
    try {
      console.log(`[GroupService] Удаление группы ${groupId}...`);
      const response = await apiClient.delete(`/admin/groups/${groupId}`);
      console.log('[GroupService] ✅ Группа успешно удалена:', response.data);
    } catch (error: any) {
      console.error('[GroupService] ❌ Ошибка удаления группы:', error);
      console.error('Детали ошибки:', error.response?.data);
      
      if (error.response?.status === 403) {
        const errorMessage = error.response?.data?.error || '';

        if (errorMessage.includes('foreign key constraint') || 
            errorMessage.includes('fk_sessions_group')) {
          throw new Error('Невозможно удалить группу, пока в ней есть события. Сначала удалите все события группы.');
        }
        
        throw new Error('Недостаточно прав для удаления группы');
      }
      
      throw new Error(error.response?.data?.message || error.response?.data?.error || 'Ошибка удаления группы');
    }
  }

  async joinGroup(groupId: number): Promise<{ joined: boolean; message: string }> {
    try {
      console.log(`[GroupService] Подача заявки в группу ${groupId}...`);
      
      const response = await apiClient.post('/groups/joinToGroup', {
        groupId: groupId
      });
      
      return response.data;
    } catch (error: any) {
      return handleApiError(error, 'Ошибка вступления в группу');
    }
  }
}

export default new GroupService();