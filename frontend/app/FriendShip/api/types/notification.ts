export interface Notification {
  id: number;
  user_id: number;
  session_id?: number | null;
  notification_type_id?: number | null;
  text: string;
  title?: string | null;
  image_url?: string | null;
  send_at: string;
  sent: boolean;
  created_at?: string;
}

export interface GroupInvite {
  id: number;
  groupId: number;
  groupName: string;
  status: string;
  createdAt: string;
}

export interface NotificationsResponse {
  notifications: Notification[];
  invites: GroupInvite[];
}

export interface MarkAsViewedRequest {
  id: number;
}

export interface MarkAsViewedResponse {
  [key: string]: string;
}

export function normalizeNotification(data: any): Notification {
  return {
    id: data.id,
    user_id: data.user_id || data.userId,
    session_id: data.session_id || data.sessionId || null,
    notification_type_id: data.notification_type_id || data.notificationTypeId || null,
    text: data.text || data.title || 'Уведомление',
    title: data.title || null,
    image_url: data.image_url || data.imageUrl || null,
    send_at: data.send_at || data.sendAt || data.created_at || new Date().toISOString(),
    sent: data.sent !== undefined ? data.sent : true,
    created_at: data.created_at || data.createdAt || new Date().toISOString(),
  };
}

export function normalizeInvite(data: any): GroupInvite {
  return {
    id: data.id,
    groupId: data.group_id || data.groupId,
    groupName: data.group_name || data.groupName || 'Группа',
    status: data.status || 'pending',
    createdAt: data.created_at || data.createdAt || new Date().toISOString(),
  };
}