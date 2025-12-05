import apiClient from '@/api/apiClient';
import { sanitizeSearchQuery } from '@/utils/searchSanitizer';
import { handleApiError } from './groupHelpers';
import { SearchGroupsParams, SearchGroupsResponse } from './groupTypes';

class GroupSearchService {

  async searchGroups(params: SearchGroupsParams): Promise<SearchGroupsResponse> {
    try {
      console.log('[GroupSearchService] Поиск групп с параметрами:', params);
      
      const response = await apiClient.get<SearchGroupsResponse>('/groups/search', {
        params: {
          name: sanitizeSearchQuery(params.name || ''),
          category: params.category || undefined,
          sort_by: params.sort_by || undefined,
          order: params.order || 'desc',
          page: Math.max(1, Math.min(params.page || 1, 100)),
        },
      });

      console.log('[GroupSearchService] ✅ Найдено групп:', response.data.groups.length);
      return response.data;
    } catch (error: any) {
      return handleApiError(error, 'Ошибка поиска групп');
    }
  }
}

export default new GroupSearchService();