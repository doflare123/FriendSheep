import apiClient from '@/api/apiClient';
import { getTokens } from '@/api/storage/tokenStorage';
import { rateLimiter } from '@/utils/rateLimiter';
import { sanitizeSearchQuery } from '@/utils/searchSanitizer';
import { validateGroupId, validateUserId } from '@/utils/validators';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

export interface GroupContact {
  name: string;
  link: string;
}

export interface CreateGroupData {
  name: string;
  description: string;
  smallDescription: string;
  city?: string;
  categories: number[];
  isPrivate: boolean;
  image: {
    uri: string;
    name: string;
    type: string;
  } | null;
  contacts?: GroupContact[];
}

export interface AdminGroup {
  id: number;
  name: string;
  small_description: string;
  image: string;
  member_count: number;
  category: string[];
  type: string;
}

export interface GroupSession {
  id: number;
  title: string;
  session_type: string;
  session_place: string; 
  city: string;
  start_time: string;
  duration: number;
  count_users_max: number;
  current_users: number;
  image_url: string;
  group_name: string;
  genres: string[];
}

export interface GroupApplication {
  id: number;
  userId: number;
  name: string;
  us: string;
  image: string;
}

export interface GroupDetailResponse {
  id: number;
  name: string;
  small_description: string;
  description: string;
  image: string;
  city: string;
  private: boolean;
  categories: string[];
  contacts: {
    name: string;
    link: string;
  }[];
  sessions: GroupSession[];
  applications: GroupApplication[];
  subscribers: GroupSubscriber[];
}

export interface PublicGroupResponse {
  id: number;
  name: string;
  description: string;
  image: string;
  city: string;
  categories: string[];
  contacts: {
    name: string;
    link: string;
  }[];
  count_members: number;
  creater: string;
  subscription: boolean;
  users: {
    id?: number;
    name: string;
    image: string;
    us: string;
  }[];
  sessions: {
    session: {
      id: number;
      title: string;
      session_type: string;
      session_place: string;
      start_time: string;
      end_time: string;
      duration: number;
      count_users_max: number;
      current_users: number;
      image_url: string;
      group_id: number;
    };
    metadata: {
      sessionID: number;
      Genres: string[];
      Location: string;
      Year: number;
      Country: string;
      AgeLimit: string;
      Notes: string;
      fields: Record<string, any>;
    };
  }[];
}

export interface UpdateGroupData {
  name?: string;
  small_description?: string;
  description?: string;
  city?: string;
  categories?: number[];
  is_private?: boolean;
  image?: string;
  contacts?: string;
}

export interface GroupRequest {
  id: number;
  userId: number;
  groupId: number;
  status: string;
  user: {
    id: number;
    name: string;
    email: string;
    image: string;
  };
  group: {
    id: number;
    name: string;
    image: string;
  };
}

export interface GroupRequestsResponse {
  requests: GroupRequest[];
}

export interface SearchGroupItem {
  id: number;
  name: string;
  description: string;
  image: string;
  category: string[];
  count: number;
  isPrivate: boolean;
  createdAt: string;
}

export interface SearchGroupsResponse {
  groups: SearchGroupItem[];
  page: number;
  total: number;
  has_more: boolean;
}

export interface SearchGroupsParams {
  name?: string;
  category?: string;
  sort_by?: 'members' | 'date' | 'category';
  order?: 'asc' | 'desc';
  page?: number;
}

export interface SimpleGroupRequest {
  id: number;
  userId: number;
  name: string;
  us: string;
  image: string;
}

export interface SimpleGroupRequestsResponse {
  requests: SimpleGroupRequest[] | null;
}

export interface GroupSubscriber {
  userId: number;
  name: string;
  image: string;
  role: string;
}

class GroupService {
  async createGroup(data: CreateGroupData): Promise<any> {
    const tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    if (!data.name?.trim() || data.name.length > 100) {
      throw new Error('Название группы должно быть от 1 до 100 символов');
    }
    
    if (!data.description?.trim() || data.description.length > 1000) {
      throw new Error('Описание не должно превышать 1000 символов');
    }
    
    if (!data.smallDescription?.trim() || data.smallDescription.length > 200) {
      throw new Error('Краткое описание не должно превышать 200 символов');
    }

    if (data.categories.length === 0) {
      throw new Error('Выберите хотя бы одну категорию');
    }

    console.log('[GroupService] Создание группы:', data.name);

    const formData = new FormData();
    
    formData.append('name', data.name.trim());
    formData.append('description', data.description.trim());
    formData.append('smallDescription', data.smallDescription.trim());
    
    if (data.city?.trim()) {
      formData.append('city', data.city.trim());
    }
    
    formData.append('isPrivate', data.isPrivate.toString());

    data.categories.forEach(categoryId => {
      formData.append('categories', categoryId.toString());
    });

    if (data.image) {
      formData.append('image', {
        uri: data.image.uri,
        name: data.image.name,
        type: data.image.type,
      } as any);
    }

    if (data.contacts && data.contacts.length > 0) {
      const sanitizedContacts = data.contacts
        .filter(contact => contact.link && contact.link.trim() !== '')
        .map(contact => {
          const name = contact.name.trim().substring(0, 50);
          let link = contact.link.trim();
          
          if (link.toLowerCase().startsWith('javascript:')) {
            throw new Error(`Недопустимый URL в контакте: ${name}`);
          }
          
          return `${name}:${link}`;
        })
        .join(', ');
      
      if (sanitizedContacts) {
        formData.append('contacts', sanitizedContacts);
      }
    }

    try {
      const response = await fetch(`${BASE_URL}/groups/createGroup`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('[GroupService] Ошибка:', response.status);
        
        try {
          const errorData = JSON.parse(errorText);
          throw new Error(errorData.message || 'Ошибка создания группы');
        } catch (parseError) {
          throw new Error(`Ошибка создания группы: ${response.status}`);
        }
      }

      const result = await response.json();
      console.log('[GroupService] Группа успешно создана');
      return result;
    } catch (error: any) {
      console.error('[GroupService] Ошибка создания группы:', error.message);
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
      const tokens = await getTokens();
      if (!tokens?.accessToken) {
        throw new Error('Пользователь не авторизован');
      }

      const filename = imageUri.split('/').pop() || 'group_image.jpg';
      const fileType = filename.split('.').pop()?.toLowerCase();

      const formData = new FormData();
      formData.append('image', {
        uri: imageUri,
        name: filename,
        type: `image/${fileType === 'jpg' ? 'jpeg' : fileType}`,
      } as any);

      console.log('Загрузка изображения группы...');

      const response = await fetch(`${BASE_URL}/admin/groups/UploadPhoto`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Ошибка загрузки изображения:', errorText);
        throw new Error('Ошибка загрузки изображения');
      }

      const result = await response.json();
      console.log('Результат загрузки изображения:', result);

      const imageUrl = result.url || result.image_url || result.image || Object.values(result)[0];
      
      if (!imageUrl || typeof imageUrl !== 'string') {
        console.error('Не удалось найти URL изображения в ответе:', result);
        throw new Error('Не удалось получить URL изображения');
      }
      
      console.log('URL загруженного изображения:', imageUrl);
      return imageUrl;
    } catch (error: any) {
      console.error('Ошибка загрузки изображения группы:', error);
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
      console.error('Ошибка обновления группы:', error);
      console.error('Детали ошибки:', error.response?.data);
      throw new Error(error.response?.data?.message || 'Ошибка обновления группы');
    }
  }

  async getGroupRequests(groupId: number): Promise<SimpleGroupRequest[]> {
    try {
      console.log(`[GroupService] Загрузка заявок для группы ${groupId}...`);
      
      const response = await apiClient.get<SimpleGroupRequestsResponse>(`/groups/requests/${groupId}`);
      
      console.log('[GroupService] ✅ Сырой ответ от API:', JSON.stringify(response.data, null, 2));
      console.log('[GroupService] ✅ Количество заявок:', response.data.requests?.length || 0);
      
      return response.data.requests || [];
    } catch (error: any) {
      console.error('[GroupService] ❌ Ошибка получения заявок:', error);
      console.error('[GroupService] ❌ Детали ошибки:', error.response?.data);
      throw new Error(error.response?.data?.message || 'Ошибка получения заявок');
    }
  }

  async approveRequest(requestId: number): Promise<void> {
    try {
      if (!Number.isInteger(requestId) || requestId <= 0) {
        throw new Error('Некорректный ID заявки');
      }

      const rateLimitKey = 'approve_request';
      if (!rateLimiter.canPerformAction(rateLimitKey, 10, 60000)) {
        throw new Error('Слишком много действий. Подождите минуту.');
      }

      console.log(`[GroupService] Одобрение заявки ${requestId}`);
      const response = await apiClient.post(`/admin/groups/requests/${requestId}/approve`);
      console.log('[GroupService] Заявка одобрена');
    } catch (error: any) {
      console.error('[GroupService] Ошибка одобрения заявки');
      throw new Error(error.response?.data?.message || 'Ошибка одобрения заявки');
    }
  }

  async rejectRequest(requestId: number): Promise<void> {
    try {
      if (!Number.isInteger(requestId) || requestId <= 0) {
        throw new Error('Некорректный ID заявки');
      }

      const rateLimitKey = 'reject_request';
      if (!rateLimiter.canPerformAction(rateLimitKey, 10, 60000)) {
        throw new Error('Слишком много действий. Подождите минуту.');
      }

      console.log(`[GroupService] Отклонение заявки ${requestId}`);
      const response = await apiClient.post(`/admin/groups/requests/${requestId}/reject`);
      console.log('[GroupService] Заявка отклонена');
    } catch (error: any) {
      console.error('[GroupService] Ошибка отклонения заявки');
      throw new Error(error.response?.data?.message || 'Ошибка отклонения заявки');
    }
  }

  async approveAllRequests(groupId: number): Promise<void> {
    try {
      console.log(`Одобрение всех заявок для группы ${groupId}...`);
      const response = await apiClient.post(`/admin/groups/requests/all/${groupId}/approveAll`);
      console.log('Все заявки одобрены:', response.data);
    } catch (error: any) {
      console.error('Ошибка одобрения всех заявок:', error);
      console.error('Детали ошибки:', error.response?.data);
      
      if (error.response?.status === 400) {
        throw new Error('Нет ожидающих заявок');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка одобрения всех заявок');
    }
  }

  async rejectAllRequests(groupId: number): Promise<void> {
    try {
      console.log(`Отклонение всех заявок для группы ${groupId}...`);
      const response = await apiClient.post(`/admin/groups/requests/all/${groupId}/rejectAll`);
      console.log('Все заявки отклонены:', response.data);
    } catch (error: any) {
      console.error('Ошибка отклонения всех заявок:', error);
      console.error('Детали ошибки:', error.response?.data);

      if (error.response?.status === 400) {
        throw new Error('Нет ожидающих заявок');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка отклонения всех заявок');
    }
  }

  async searchGroups(params: SearchGroupsParams): Promise<SearchGroupsResponse> {
    try {
      console.log('[GroupService] Поиск групп');
      
      const response = await apiClient.get<SearchGroupsResponse>('/groups/search', {
        params: {
          name: sanitizeSearchQuery(params.name || ''),
          category: params.category || undefined,
          sort_by: params.sort_by || undefined,
          order: params.order || 'desc',
          page: Math.max(1, Math.min(params.page || 1, 100)),
        },
      });

      return response.data;
    } catch (error: any) {
      console.error('[GroupService] Ошибка поиска групп');
      throw new Error(error.response?.data?.message || 'Ошибка поиска групп');
    }
  }

  async joinGroup(groupId: number): Promise<{ joined: boolean; message: string }> {
    try {
      console.log(`[GroupService] Подача заявки в группу ${groupId}...`);
      
      const response = await apiClient.post('/groups/joinToGroup', {
        groupId: groupId
      });
      
      console.log('[GroupService] Ответ сервера:', response.data);
      
      return response.data;
    } catch (error: any) {
      console.error('[GroupService] Ошибка вступления в группу:', error);
      console.error('Детали ошибки:', error.response?.data);
      throw new Error(error.response?.data?.message || 'Ошибка вступления в группу');
    }
  }

  async leaveGroup(groupId: number): Promise<void> {
  try {
    console.log(`[GroupService] 🚪 Выход из группы ${groupId}...`);

    const response = await apiClient.delete(`/groups/${groupId}/leave`);
    
    console.log('[GroupService] ✅ Группа покинута:', response.data);
  } catch (error: any) {
    console.error('[GroupService] ❌ Ошибка выхода из группы:', error);
    
    if (error.response?.status === 400) {
      const errorMsg = error.response?.data?.error || '';
      if (errorMsg.includes('единственный админ')) {
        throw new Error('Вы не можете покинуть группу, так как являетесь единственным администратором');
      }
      throw new Error('Вы не состоите в этой группе');
    }
    if (error.response?.status === 401) {
      throw new Error('Необходимо войти в систему');
    }
    
    throw new Error(error.response?.data?.error || error.response?.data?.message || 'Ошибка выхода из группы');
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

  async removeMember(groupId: number, userId: number): Promise<void> {
    try {
      console.log(`[GroupService] Удаление участника ${userId} из группы ${groupId}...`);
      const response = await apiClient.delete(`/admin/groups/${groupId}/members/${userId}`);
      console.log('[GroupService] ✅ Участник удалён:', response.data);
    } catch (error: any) {
      console.error('[GroupService] ❌ Ошибка удаления участника:', error);
      console.error('Детали ошибки:', error.response?.data);
      
      if (error.response?.status === 403) {
        throw new Error('Недостаточно прав для удаления участника');
      }
      if (error.response?.status === 404) {
        throw new Error('Участник не найден в группе');
      }
      
      throw new Error(error.response?.data?.message || error.response?.data?.error || 'Ошибка удаления участника');
    }
  }

  async respondToInvite(inviteId: string, action: 'accepted' | 'rejected'): Promise<void> {
    try {
      console.log('[GroupService] 📨 Ответ на приглашение:', { inviteId, action });

      const endpoint = action === 'accepted' 
        ? `/users/invites/${inviteId}/approve`
        : `/users/invites/${inviteId}/reject`;
      
      await apiClient.put(endpoint);
      
      console.log('[GroupService] ✅ Приглашение обработано:', action);
    } catch (error: any) {
      console.error('[GroupService] ❌ Ошибка ответа на приглашение:', error);
      
      if (error.response?.status === 400) {
        throw new Error('Приглашение уже обработано или недоступно');
      }
      if (error.response?.status === 401) {
        throw new Error('Необходимо войти в систему');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка обработки приглашения');
    }
  }

  async sendInviteToUser(groupId: number, userId: number): Promise<{ joined: boolean; message: string }> {
    try {
      const validGroupId = validateGroupId(groupId);
      const validUserId = validateUserId(userId);

      const rateLimitKey = `invite_user_${validGroupId}`;
      if (!rateLimiter.canPerformAction(rateLimitKey, 10, 60000)) {
        throw new Error('Слишком много приглашений. Подождите минуту.');
      }

      console.log('[GroupService] Отправка приглашения');
      
      const response = await apiClient.post<{ joined: boolean; message: string }>(
        '/admin/groups/requestsForUser',
        null,
        {
          params: {
            group_id: validGroupId,
            user_id: validUserId,
          },
        }
      );
      
      return response.data;
    } catch (error: any) {
      console.error('[GroupService] Ошибка отправки приглашения');
      
      if (error.response?.status === 403) {
        throw new Error('У вас нет прав администратора этой группы');
      }
      if (error.response?.status === 404) {
        throw new Error('Пользователь или группа не найдены');
      }
      if (error.response?.status === 400) {
        const errorMsg = error.response?.data?.message || '';
        if (errorMsg.includes('already')) {
          throw new Error('Пользователь уже состоит в группе');
        }
        throw new Error(errorMsg || 'Некорректные параметры запроса');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка отправки приглашения');
    }
  }
}

export default new GroupService();