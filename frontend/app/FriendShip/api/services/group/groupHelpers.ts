import { CreateGroupData, GroupContact } from './groupTypes';

export const validateCreateGroupData = (data: CreateGroupData): void => {
  if (!data.name?.trim() || data.name.length > 100) {
    throw new Error('Название группы должно быть от 1 до 100 символов');
  }
  
  if (!data.description?.trim() || data.description.length > 1000) {
    throw new Error('Описание не должно превышать 1000 символов');
  }
  
  if (!data.smallDescription?.trim() || data.smallDescription.length > 200) {
    throw new Error('Краткое описание не должно превышать 200 символов');
  }

  if (data.categories.length === 0) {
    throw new Error('Выберите хотя бы одну категорию');
  }
};

export const createGroupFormData = (data: CreateGroupData): FormData => {
  const formData = new FormData();
  
  formData.append('name', data.name.trim());
  formData.append('description', data.description.trim());
  formData.append('smallDescription', data.smallDescription.trim());
  
  if (data.city?.trim()) {
    formData.append('city', data.city.trim());
  }
  
  formData.append('isPrivate', data.isPrivate.toString());

  data.categories.forEach(categoryId => {
    formData.append('categories', categoryId.toString());
  });

  if (data.image) {
    formData.append('image', {
      uri: data.image.uri,
      name: data.image.name,
      type: data.image.type,
    } as any);
  }

  if (data.contacts && data.contacts.length > 0) {
    const sanitizedContacts = sanitizeGroupContacts(data.contacts);
    if (sanitizedContacts) {
      formData.append('contacts', sanitizedContacts);
    }
  }

  return formData;
};

export const sanitizeGroupContacts = (contacts: GroupContact[]): string => {
  const sanitizedContacts = contacts
    .filter(contact => contact.link && contact.link.trim() !== '')
    .map(contact => {
      const name = contact.name.trim().substring(0, 50);
      let link = contact.link.trim();
      
      if (link.toLowerCase().startsWith('javascript:')) {
        throw new Error(`Недопустимый URL в контакте: ${name}`);
      }
      
      return `${name}:${link}`;
    })
    .join(', ');
  
  return sanitizedContacts;
};

export const createPhotoFormData = (imageUri: string): FormData => {
  const filename = imageUri.split('/').pop() || 'group_image.jpg';
  const fileType = filename.split('.').pop()?.toLowerCase();

  const formData = new FormData();
  formData.append('image', {
    uri: imageUri,
    name: filename,
    type: `image/${fileType === 'jpg' ? 'jpeg' : fileType}`,
  } as any);

  return formData;
};

export const extractImageUrl = (result: any): string => {
  const imageUrl = result.url || result.image_url || result.image || Object.values(result)[0];
  
  if (!imageUrl || typeof imageUrl !== 'string') {
    console.error('Не удалось найти URL изображения в ответе:', result);
    throw new Error('Не удалось получить URL изображения');
  }
  
  return imageUrl;
};

export const handleApiError = (error: any, defaultMessage: string): never => {
  console.error(`[GroupService] ${defaultMessage}:`, error);
  console.error('Детали ошибки:', error.response?.data);
  throw new Error(error.response?.data?.message || error.response?.data?.error || defaultMessage);
};

export const validateId = (id: number, name: string = 'ID'): void => {
  if (!Number.isInteger(id) || id <= 0) {
    throw new Error(`Некорректный ${name}`);
  }
};