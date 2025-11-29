export const validateUserId = (userId: any): number => {
  const parsed = typeof userId === 'string' ? parseInt(userId) : userId;
  
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error('Некорректный ID пользователя');
  }
  
  return parsed;
};

export const validateGroupId = (groupId: any): number => {
  const parsed = typeof groupId === 'string' ? parseInt(groupId) : groupId;
  
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error('Некорректный ID группы');
  }
  
  return parsed;
};

export const validateSessionId = (sessionId: any): number => {
  const parsed = typeof sessionId === 'string' ? parseInt(sessionId) : sessionId;
  
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error('Некорректный ID события');
  }
  
  return parsed;
};

export const validateUserStats = (stats: any) => {
  return {
    spent_time: Math.max(0, Math.min(stats?.spent_time || 0, 100000)),
    count_sessions: Math.max(0, Math.min(stats?.count_sessions || 0, 10000)),
    most_pop_day: stats?.most_pop_day || 'Не определён',
  };
};

export const validateGenres = (genres: string[] | undefined): string[] => {
  if (!Array.isArray(genres)) return [];
  
  return genres
    .filter(g => typeof g === 'string')
    .slice(0, 10)
    .map(g => g.substring(0, 50));
};

export const validateNotificationData = (notification: any): any => {
  return {
    ...notification,
    title: notification.title?.substring(0, 100) || '',
    description: notification.description?.substring(0, 500) || '',
    id: typeof notification.id === 'number' ? notification.id : 0,
    groupId: typeof notification.groupId === 'number' ? notification.groupId : undefined,
    type: notification.type || 'unknown',
    viewed: typeof notification.viewed === 'boolean' ? notification.viewed : false,
    createdAt: notification.createdAt || new Date().toISOString(),
  };
};