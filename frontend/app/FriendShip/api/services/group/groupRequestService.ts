import apiClient from '@/api/apiClient';
import { rateLimiter } from '@/utils/rateLimiter';
import { handleApiError, validateId } from './groupHelpers';
import { SimpleGroupRequest, SimpleGroupRequestsResponse } from './groupTypes';

class GroupRequestService {

  async getGroupRequests(groupId: number): Promise<SimpleGroupRequest[]> {
    try {
      console.log(`[GroupRequestService] Загрузка заявок для группы ${groupId}...`);
      
      const response = await apiClient.get<SimpleGroupRequestsResponse>(`/groups/requests/${groupId}`);
      
      console.log('[GroupRequestService] ✅ Сырой ответ от API:', JSON.stringify(response.data, null, 2));
      console.log('[GroupRequestService] ✅ Количество заявок:', response.data.requests?.length || 0);
      
      return response.data.requests || [];
    } catch (error: any) {
      return handleApiError(error, 'Ошибка получения заявок');
    }
  }

  async approveRequest(requestId: number): Promise<void> {
    try {
      validateId(requestId, 'ID заявки');

      const rateLimitKey = 'approve_request';
      if (!rateLimiter.canPerformAction(rateLimitKey, 10, 60000)) {
        throw new Error('Слишком много действий. Подождите минуту.');
      }

      console.log(`[GroupRequestService] Одобрение заявки ${requestId}`);
      await apiClient.post(`/admin/groups/requests/${requestId}/approve`);
      console.log('[GroupRequestService] Заявка одобрена');
    } catch (error: any) {
      console.error('[GroupRequestService] Ошибка одобрения заявки');
      throw new Error(error.response?.data?.message || 'Ошибка одобрения заявки');
    }
  }

  async rejectRequest(requestId: number): Promise<void> {
    try {
      validateId(requestId, 'ID заявки');

      const rateLimitKey = 'reject_request';
      if (!rateLimiter.canPerformAction(rateLimitKey, 10, 60000)) {
        throw new Error('Слишком много действий. Подождите минуту.');
      }

      console.log(`[GroupRequestService] Отклонение заявки ${requestId}`);
      await apiClient.post(`/admin/groups/requests/${requestId}/reject`);
      console.log('[GroupRequestService] Заявка отклонена');
    } catch (error: any) {
      console.error('[GroupRequestService] Ошибка отклонения заявки');
      throw new Error(error.response?.data?.message || 'Ошибка отклонения заявки');
    }
  }

  async approveAllRequests(groupId: number): Promise<void> {
    try {
      console.log(`[GroupRequestService] Одобрение всех заявок для группы ${groupId}...`);
      const response = await apiClient.post(`/admin/groups/requests/all/${groupId}/approveAll`);
      console.log('[GroupRequestService] Все заявки одобрены:', response.data);
    } catch (error: any) {
      console.error('[GroupRequestService] Ошибка одобрения всех заявок:', error);
      console.error('Детали ошибки:', error.response?.data);
      
      if (error.response?.status === 400) {
        throw new Error('Нет ожидающих заявок');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка одобрения всех заявок');
    }
  }

  async rejectAllRequests(groupId: number): Promise<void> {
    try {
      console.log(`[GroupRequestService] Отклонение всех заявок для группы ${groupId}...`);
      const response = await apiClient.post(`/admin/groups/requests/all/${groupId}/rejectAll`);
      console.log('[GroupRequestService] Все заявки отклонены:', response.data);
    } catch (error: any) {
      console.error('[GroupRequestService] Ошибка отклонения всех заявок:', error);
      console.error('Детали ошибки:', error.response?.data);

      if (error.response?.status === 400) {
        throw new Error('Нет ожидающих заявок');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка отклонения всех заявок');
    }
  }
}

export default new GroupRequestService();