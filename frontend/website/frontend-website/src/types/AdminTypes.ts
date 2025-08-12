// src/types/AdminTypes.ts

export interface AdminMenuSection {
  id: string;
  title: string;
  component?: React.ComponentType<any>;
}

export interface GroupFormData {
  name: string;
  shortDescription: string;
  description: string;
  city: string;
  isPrivate: boolean;
  categories: string[];
  socialContacts: { name: string; link: string }[];
  image?: File;
  imagePreview?: string;
}