'use client';

import React, { useState } from 'react';
import AdminSidebar from './AdminSidebar';
import CreateGroupForm from './CreateGroupForm';
import RequestsManagementComponent from './RequestsManagementComponent';
import EventsManagementComponent from './EventsManagementComponent';
import MembersManagement from './MembersManagement';
import { AdminMenuSection } from '../../types/AdminTypes';
import { convertCategoriesToIds, convertIdsToCategories, convertSocialContactsToString, getAccesToken } from '../../Constants';
import { GroupData } from '../../types/Groups';
import styles from '../../styles/Groups/admin/AdminPage.module.css';
import { editGroup } from '../../api/groups/edit_group';
import LoadingIndicator from '@/components/LoadingIndicator';
import { showNotification } from '@/utils';
import { getImage } from '@/api/getImage';
import { useRouter } from 'next/navigation';
import { delGroup } from '@/api/groups/delGroup';

// Компонент-заглушка для пустых разделов
const EmptySection: React.FC<{ title: string }> = ({ title }) => (
  <div className={styles.emptyState}>
    <div className={styles.emptyIcon}>📋</div>
    <div>Раздел "{title}" в разработке</div>
  </div>
);

// Компонент для отображения основной информации группы
const GroupInfoSection: React.FC<{ 
  groupData?: GroupData; 
  groupId?: string;
  onGroupDataUpdate?: (updatedData: Partial<GroupData>) => void;
}> = ({ groupData, groupId, onGroupDataUpdate }) => {
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const handleFormSubmit = async (formData: any) => {
    if (!groupId) {
      showNotification(400, 'ID группы не найден');
      return;
    }

    setIsLoading(true);

    try {
      const accessToken = await getAccesToken(router);
      
      // Формируем начальные данные для сравнения
      const initialData = {
        name: groupData?.name,
        description: groupData?.description,
        small_description: groupData?.small_description || '',
        city: groupData?.city,
        categories: groupData?.categories,
        isPrivate: groupData?.private || false,
        image: groupData?.image,
        contacts: groupData?.contacts?.map(contact => ({
          name: contact.name,
          link: contact.link
        })) || [],
      };

      // Объект для хранения изменённых полей
      const changedFields: any = {};

      // Сравниваем name
      if (formData.name !== initialData.name) {
        changedFields.name = formData.name;
      }

      // Сравниваем description
      if (formData.description !== initialData.description) {
        changedFields.description = formData.description;
      }

      // Сравниваем shortDescription (переименовываем в small_description)
      if (formData.shortDescription !== initialData.small_description) {
        changedFields.small_description = formData.shortDescription;
      }

      // Сравниваем city
      if (formData.city !== initialData.city) {
        changedFields.city = formData.city;
      }

      // Сравниваем isPrivate
      if (formData.isPrivate !== initialData.isPrivate) {
        changedFields.isPrivate = formData.isPrivate;
      }

      // Сравниваем categories
      const formCategories = convertCategoriesToIds(formData.categories);
      const initialCategories = initialData.categories || [];
      if (JSON.stringify(formCategories.sort()) !== JSON.stringify(initialCategories.sort())) {
        changedFields.categories = formCategories;
      }

      // Сравниваем контакты
      const formContactsString = convertSocialContactsToString(formData.socialContacts);
      const initialContactsString = convertSocialContactsToString(initialData.contacts);
      if (formContactsString !== initialContactsString) {
        changedFields.contacts = formContactsString;
      }

      // Обрабатываем изображение
      if (formData.image && formData.image instanceof File) {
        // Загружен новый файл - загружаем его и получаем строку
        const imageString = await getImage(accessToken, formData.image);
        changedFields.image = imageString;
      } else if (formData.image === null && initialData.image) {
        // Изображение удалено
        changedFields.image = null;
      }

      // Проверяем, есть ли хоть одно изменённое поле
      if (Object.keys(changedFields).length === 0) {
        showNotification(200, 'Нет изменений для сохранения');
        setIsLoading(false);
        return;
      }

      // Отправляем только изменённые поля
      await editGroup(
        accessToken,
        parseInt(groupId),
        changedFields.name,
        changedFields.description,
        changedFields.small_description,
        changedFields.city,
        changedFields.categories,
        changedFields.isPrivate,
        changedFields.image,
        changedFields.contacts
      );

      // Обновляем локальные данные группы
      if (onGroupDataUpdate) {
        const updatedData: Partial<GroupData> = {};
        
        console.log("ZZZ", changedFields);

        if (changedFields.name !== undefined) updatedData.name = changedFields.name;
        if (changedFields.description !== undefined) updatedData.description = changedFields.description;
        if (changedFields.small_description !== undefined) updatedData.small_description = changedFields.small_description;
        if (changedFields.city !== undefined) updatedData.city = changedFields.city;
        if (changedFields.isPrivate !== undefined) updatedData.private = changedFields.isPrivate;
        if (changedFields.categories !== undefined) updatedData.categories = changedFields.categories;
        if (changedFields.image !== undefined && changedFields.image !== null) updatedData.image = changedFields.image;
        if (changedFields.contacts !== undefined) {
          // Парсим строку контактов обратно в массив объектов
          if (changedFields.contacts.trim()) {
            updatedData.contacts = changedFields.contacts.split(', ').map((contact: string) => {
              const [name, link] = contact.split(':');
              return { name: name.trim(), link: link.trim() };
            });
          } else {
            updatedData.contacts = [];
          }
        }
        
        onGroupDataUpdate(updatedData);
      }

      showNotification(200, 'Данные группы успешно сохранены!');
    } catch (error: any) {
      console.error('Ошибка при сохранении:', error);
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Произошла ошибка при сохранении';
      showNotification(statusCode, errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!groupId) {
      showNotification(400, 'ID группы не найден');
      return;
    }

    setIsLoading(true);

    try {
      const accessToken = await getAccesToken(router);
      
      await delGroup(accessToken, parseInt(groupId));
      
      showNotification(200, 'Группа успешно удалена');
      
      // Перенаправляем на страницу групп или главную
      router.push('/groups'); // Или куда нужно
    } catch (error: any) {
      console.error('Ошибка при удалении группы:', error);
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Произошла ошибка при удалении группы';
      showNotification(statusCode, errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  // Преобразуем данные группы в формат для формы
  const initialFormData = groupData ? {
    name: groupData.name,
    shortDescription: groupData.small_description || '',
    description: groupData.description,
    city: groupData.city,
    isPrivate: groupData.private || false,
    categories: groupData.categories || [],
    socialContacts: groupData.contacts?.map(contact => ({
      name: contact.name,
      link: contact.link
    })) || [],
    imagePreview: groupData.image
  } : undefined;

  return (
    <div>
      {isLoading ? (
        <LoadingIndicator text="Обработка..." />
      ) : (
        <CreateGroupForm 
          onSubmit={handleFormSubmit}
          onDelete={handleDelete} // Добавь этот проп
          initialData={initialFormData}
          showTitle={false}
          isLoading={isLoading}
          isEditMode={true} // Добавь этот проп
          groupName={groupData?.name} // Добавь этот проп
        />
      )}
    </div>
  );
};

interface GroupAdminComponentProps {
  groupId?: string;
  groupData?: GroupData;
  onGroupDataUpdate?: (updatedData: Partial<GroupData>) => void;
}

const GroupAdminComponent: React.FC<GroupAdminComponentProps> = ({ 
  groupId, 
  groupData,
  onGroupDataUpdate 
}) => {
  const [activeSection, setActiveSection] = useState('info');
  
  const menuSections: AdminMenuSection[] = [
    {
      id: 'info',
      title: 'Основная информация о группе',
      component: GroupInfoSection
    },
    {
      id: 'requests',
      title: 'Управление заявками',
      component: RequestsManagementComponent
    },
    {
      id: 'members',
      title: 'Управление участниками',
      component: MembersManagement
    },
    {
      id: 'events',
      title: 'Управление событиями',
      component: EventsManagementComponent
    }
  ];

  const currentSection = menuSections.find(section => section.id === activeSection);
  const CurrentComponent = currentSection?.component;

  const getSectionTitle = () => {
    switch (activeSection) {
      case 'info':
        return 'Основная информация';
      case 'requests':
        return 'Управление заявками';
      case 'members':
        return 'Управление участниками';
      case 'events':
        return 'Управление событиями';
      default:
        return 'Администрирование группы';
    }
  };

  return (
    <div className="bgPage">
      <div className={styles.adminPageWrapper}>
        <div className={styles.adminPage}>
          <AdminSidebar
            sections={menuSections}
            activeSection={activeSection}
            onSectionChange={setActiveSection}
          />
          
          <div className={styles.mainContent}>
            <div className={styles.contentHeader}>
              <h2>{getSectionTitle()}</h2>
            </div>
            
            <div className={styles.contentBody}>
              {CurrentComponent ? (
                <CurrentComponent 
                  groupData={groupData} 
                  groupId={groupId}
                  onGroupDataUpdate={onGroupDataUpdate}
                  useMockData={false}
                />
              ) : (
                <EmptySection title={currentSection?.title || ''} />
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default GroupAdminComponent;