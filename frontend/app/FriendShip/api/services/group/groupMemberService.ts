import apiClient from '@/api/apiClient';
import { rateLimiter } from '@/utils/rateLimiter';
import { validateGroupId, validateUserId } from '@/utils/validators';

class GroupMemberService {

  async leaveGroup(groupId: number): Promise<void> {
    try {
      console.log(`[GroupMemberService] 🚪 Выход из группы ${groupId}...`);

      const response = await apiClient.delete(`/groups/${groupId}/leave`);
      
      console.log('[GroupMemberService] ✅ Группа покинута:', response.data);
    } catch (error: any) {
      console.error('[GroupMemberService] ❌ Ошибка выхода из группы:', error);
      
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

  async removeMember(groupId: number, userId: number): Promise<void> {
    try {
      console.log(`[GroupMemberService] Удаление участника ${userId} из группы ${groupId}...`);
      const response = await apiClient.delete(`/admin/groups/${groupId}/members/${userId}`);
      console.log('[GroupMemberService] ✅ Участник удалён:', response.data);
    } catch (error: any) {
      console.error('[GroupMemberService] ❌ Ошибка удаления участника:', error);
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
      console.log('[GroupMemberService] 📨 Ответ на приглашение:', { inviteId, action });

      const endpoint = action === 'accepted' 
        ? `/users/invites/${inviteId}/approve`
        : `/users/invites/${inviteId}/reject`;
      
      await apiClient.put(endpoint);
      
      console.log('[GroupMemberService] ✅ Приглашение обработано:', action);
    } catch (error: any) {
      console.error('[GroupMemberService] ❌ Ошибка ответа на приглашение:', error);
      
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

      console.log('[GroupMemberService] Отправка приглашения');
      
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
        if (error.response) {
          // ✅ Детально обрабатываем ошибки сервера
          const status = error.response.status;
          const data = error.response.data;
          
          // Конкретные статусы
          if (status === 403) throw new Error('...');
          if (status === 404) throw new Error('...');
          if (status === 400) throw new Error('...');
          
          // ✅ Для всех остальных - понятное сообщение
          throw new Error(data?.message || data?.error || `Ошибка сервера (${status})`);
          
        } else if (error.request) {
          // ✅ Нет ответа от сервера
          throw new Error('Сервер не отвечает. Проверьте подключение к интернету.');
          
        } else {
          // ✅ Ошибка настройки запроса
          throw new Error(error.message || 'Неизвестная ошибка');
        }
      }
  }
}

export default new GroupMemberService();