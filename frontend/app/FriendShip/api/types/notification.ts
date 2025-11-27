export interface Notification {
  id: number;
  text: string;
  type: string;
  sendAt: string;
  viewed: boolean;
  sent: boolean;
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