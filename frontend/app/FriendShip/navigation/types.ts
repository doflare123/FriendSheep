export type RootStackParamList = {
  Register: undefined;
  Login: undefined;
  Confirm: undefined;
  Done: undefined;
  MainPage: { searchQuery?: string } | undefined;
  CategoryPage: {
    category: 'movie' | 'game' | 'table_game' | 'other' | 'popular' | 'new';
    title: string;
    imageSource: any;
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
  ProfilePage: { userId?: string } | undefined;
  UserSearchPage: undefined;
  SettingsPage: undefined;
};
