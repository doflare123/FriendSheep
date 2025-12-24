import apiClient from '@/api/apiClient';
import { sanitizeSearchQuery } from '@/utils/searchSanitizer';
import { handleApiError } from './groupHelpers';
import { SearchGroupsParams, SearchGroupsResponse } from './groupTypes';

class GroupSearchService {
  async searchGroups(params: SearchGroupsParams): Promise<SearchGroupsResponse> {
    try {
      const response = await apiClient.get<SearchGroupsResponse>('/groups/search', {
        params: {
          name: sanitizeSearchQuery(params.name || ''),
          category: params.category || undefined,
          sort_by: params.sort_by || undefined,
          order: params.order || 'desc',
          page: Math.max(1, Math.min(params.page || 1, 100)),
        },
      });

      const normalizedGroups = response.data.groups
        .map(group => ({
          ...group,
          category: Array.isArray(group.category) ? group.category : [],
        }))
        .filter(group => group.category.length > 0);

      if (response.data.groups.length > normalizedGroups.length) {
        console.warn(
          `[GroupSearchService] Исключено ${response.data.groups.length - normalizedGroups.length} групп без категорий`
        );
      }

      return {
        ...response.data,
        groups: normalizedGroups,
        total: response.data.total,
      };
    } catch (error: any) {
      return handleApiError(error, 'Ошибка поиска групп');
    }
  }
}

export default new GroupSearchService();