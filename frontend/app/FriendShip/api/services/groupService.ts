import apiClient from '@/api/apiClient';
import { getTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL, LOCAL_IP } from '@env';

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
    name: string;
    image: string;
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
      genres: string[];
      location: string;
      year: number;
      country: string;
      ageLimit: string;
      notes: string;
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

class GroupService {
  async createGroup(groupData: CreateGroupData): Promise<any> {
    const tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    console.log('=== Создание группы ===');
    console.log('Данные группы:', {
      name: groupData.name,
      description: groupData.description,
      smallDescription: groupData.smallDescription,
      city: groupData.city,
      isPrivate: groupData.isPrivate,
      categories: groupData.categories,
      hasImage: !!groupData.image,
      contactsCount: groupData.contacts?.length || 0,
    });

    const formData = new FormData();
    
    formData.append('name', groupData.name);
    formData.append('description', groupData.description);
    formData.append('smallDescription', groupData.smallDescription);
    
    if (groupData.city) {
      formData.append('city', groupData.city);
    }
    
    formData.append('isPrivate', groupData.isPrivate.toString());

    groupData.categories.forEach(categoryId => {
      formData.append('categories', categoryId.toString());
    });

    console.log('Категории для отправки:', groupData.categories);

    if (groupData.image) {
      formData.append('image', {
        uri: groupData.image.uri,
        name: groupData.image.name,
        type: groupData.image.type,
      } as any);
      console.log('Изображение добавлено:', groupData.image.name);
    }

    if (groupData.contacts && groupData.contacts.length > 0) {
      const contactsString = groupData.contacts
        .filter(contact => contact.link && contact.link.trim() !== '')
        .map(contact => {
          const name = contact.name.trim();
          return `${name}:${contact.link.trim()}`;
        })
        .join(', ');
      
      if (contactsString) {
        formData.append('contacts', contactsString);
        console.log('Контакты для отправки:', contactsString);
      }
    }

    console.log('Отправка запроса на:', `${BASE_URL}/groups/createGroup`);

    try {
      const response = await fetch(`${BASE_URL}/groups/createGroup`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      console.log('Статус ответа:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Ошибка от сервера (текст):', errorText);
        
        try {
          const errorData = JSON.parse(errorText);
          console.error('Ошибка от сервера (JSON):', errorData);
          throw new Error(errorData.message || 'Ошибка создания группы');
        } catch (parseError) {
          throw new Error(`Ошибка создания группы: ${response.status} ${response.statusText}`);
        }
      }

      const result = await response.json();
      console.log('Группа успешно создана:', result);
      return result;
    } catch (error: any) {
      console.error('Ошибка при создании группы:', error);
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
      throw new Error(error.response?.data?.message || 'Ошибка получения информации о группе');
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

  async getGroupRequests(groupId: number): Promise<GroupRequest[]> {
    try {
      console.log(`Загрузка заявок для группы ${groupId}...`);
      const response = await apiClient.get<GroupRequestsResponse>(`/groups/requests/${groupId}`);
      
      console.log('Заявки получены:', response.data);
      
      return response.data.requests || [];
    } catch (error: any) {
      console.error('Ошибка получения заявок:', error);
      console.error('Детали ошибки:', error.response?.data);
      throw new Error(error.response?.data?.message || 'Ошибка получения заявок');
    }
  }

  async approveRequest(requestId: number): Promise<void> {
  try {
    console.log(`Одобрение заявки ${requestId}...`);
    const response = await apiClient.post(`/admin/groups/requests/${requestId}/approve`);
    console.log('Заявка одобрена:', response.data);
  } catch (error: any) {
    console.error('Ошибка одобрения заявки:', error);
    console.error('Детали ошибки:', error.response?.data);
    throw new Error(error.response?.data?.message || 'Ошибка одобрения заявки');
  }
}

  async rejectRequest(requestId: number): Promise<void> {
    try {
      console.log(`Отклонение заявки ${requestId}...`);
      const response = await apiClient.post(`/admin/groups/requests/${requestId}/reject`);
      console.log('Заявка отклонена:', response.data);
    } catch (error: any) {
      console.error('Ошибка отклонения заявки:', error);
      console.error('Детали ошибки:', error.response?.data);
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
      console.log('[GroupService] Поиск групп с параметрами:', params);
      
      const response = await apiClient.get<SearchGroupsResponse>('/groups/search', {
        params: {
          name: params.name || undefined,
          category: params.category || undefined,
          sort_by: params.sort_by || undefined,
          order: params.order || 'desc',
          page: params.page || 1,
        },
      });
      
      console.log('[GroupService] Результаты поиска:', {
        total: response.data.total,
        page: response.data.page,
        groupsCount: response.data.groups.length,
        has_more: response.data.has_more,
      });

      const normalizedGroups = response.data.groups.map(group => ({
        ...group,
        image: group.image?.includes('localhost')
          ? group.image.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080')
          : group.image,
      }));
      
      return {
        ...response.data,
        groups: normalizedGroups,
      };
    } catch (error: any) {
      console.error('[GroupService] Ошибка поиска групп:', error);
      console.error('Детали ошибки:', error.response?.data);
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
      console.log(`[GroupService] Выход из группы ${groupId}...`);
      const response = await apiClient.post(`/groups/${groupId}/leave`);
      console.log('[GroupService] Группа покинута:', response.data);
    } catch (error: any) {
      console.error('[GroupService] Ошибка выхода из группы:', error);
      throw new Error(error.response?.data?.message || 'Ошибка выхода из группы');
    }
  }
}

export default new GroupService();