export interface RequestData {
  id: number;
  userId: number;
  groupId: number;
  status: 'pending' | 'accepted' | 'rejected';
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