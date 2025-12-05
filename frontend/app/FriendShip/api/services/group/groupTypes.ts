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