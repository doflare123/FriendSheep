export interface Notification {
  id: number;
  text: string;
  type: string;
  sendAt: string;
  viewed: boolean;
  sent: boolean;
  groupId?: number;
  sessionId?: number;
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

export interface UnreadNotificationsResponse {
  has_unread: boolean;
}