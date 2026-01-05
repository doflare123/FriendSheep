export type RootStackParamList = {
  Register: undefined;
  Login: undefined;
  ForgotPassword: undefined;
  ResetPassword: undefined;
  Confirm: undefined;
  Done: undefined;
  MainPage: { searchQuery?: string } | undefined;
  CategoryPage: {
    category: 'movie' | 'game' | 'table_game' | 'other' | 'popular' | 'new';
    title: string;
  };
  GroupsPage: undefined;
  GroupPage: {
    groupId: string;
    mode?: 'manage' | 'view';
  };
  GroupManagePage: {
    groupId: string;
  };
  GroupSearchPage: undefined;
  ProfilePage: {
    userId?: string 
    openNotifications?: boolean; 
    openEventId?: string;
  } | undefined;     
  UserSearchPage: undefined;
  SettingsPage: undefined;
  AllEventsPage: {
    mode: 'user' | 'group';
    groupId?: string;
    userId?: string;
  };
  AllGroupsPage: undefined;
  AboutPage: undefined;
};
