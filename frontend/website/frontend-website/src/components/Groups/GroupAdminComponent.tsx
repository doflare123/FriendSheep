// src/components/Groups/GroupAdminComponent.tsx

'use client';

import React, { useState } from 'react';
import AdminSidebar from './AdminSidebar';
import CreateGroupForm from './CreateGroupForm';
import RequestsManagementComponent from './RequestsManagementComponent';
import { AdminMenuSection } from '../../types/AdminTypes';
import { convertCategoriesToIds, convertSocialContactsToString } from '../../Constants';
import { GroupData } from '../../types/Groups';
import styles from '../../styles/Groups/admin/AdminPage.module.css';
import {editGroup} from '../../api/edit_group';

// Компонент-заглушка для пустых разделов
const EmptySection: React.FC<{ title: string }> = ({ title }) => (
  <div className={styles.emptyState}>
    <div className={styles.emptyIcon}>📋</div>
    <div>Раздел "{title}" в разработке</div>
  </div>
);

// Компонент для отображения основной информации группы
const GroupInfoSection: React.FC<{ groupData?: GroupData; groupId?: string }> = ({ groupData, groupId }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const handleFormSubmit = async (formData: any) => {
    if (!groupId) {
      setSaveMessage({ type: 'error', text: 'ID группы не найден' });
      return;
    }

    setIsLoading(true);
    setSaveMessage(null);

    try {
      // Получаем токен авторизации (здесь нужно заменить на реальный способ получения токена)
      const accessToken = localStorage.getItem('access_token') || '';
      
      // Преобразуем категории в числа
      const categoriesNumbers = convertCategoriesToIds(formData.categories);
      
      // Преобразуем контакты в строку
      const contactsString = convertSocialContactsToString(formData.socialContacts);

      await editGroup(
        formData.name,
        formData.description,
        formData.shortDescription,
        formData.city,
        categoriesNumbers,
        formData.isPrivate,
        formData.image,
        contactsString,
        accessToken,
        parseInt(groupId)
      );

      setSaveMessage({ type: 'success', text: 'Данные группы успешно сохранены!' });
    } catch (error: any) {
      console.error('Ошибка при сохранении:', error);
      setSaveMessage({ 
        type: 'error', 
        text: error.response?.data?.message || 'Произошла ошибка при сохранении' 
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Преобразуем данные группы в формат для формы
  const initialFormData = groupData ? {
    name: groupData.name,
    shortDescription: '', // У GroupData нет этого поля, можно добавить или оставить пустым
    description: groupData.description,
    city: groupData.city,
    isPrivate: false, // У GroupData нет этого поля, можно добавить или задать по умолчанию
    categories: groupData.categories,
    socialContacts: groupData.contacts?.map(contact => ({
      name: contact.name,
      link: contact.link
    })) || [],
    imagePreview: groupData.image
  } : undefined;

  return (
    <div>
      {saveMessage && (
        <div style={{
          padding: '12px 16px',
          marginBottom: '20px',
          borderRadius: '8px',
          backgroundColor: saveMessage.type === 'success' ? '#d4edda' : '#f8d7da',
          border: `1px solid ${saveMessage.type === 'success' ? '#c3e6cb' : '#f5c6cb'}`,
          color: saveMessage.type === 'success' ? '#155724' : '#721c24'
        }}>
          {saveMessage.text}
        </div>
      )}
      
      <CreateGroupForm 
        onSubmit={handleFormSubmit}
        initialData={initialFormData}
        showTitle={false}
        isLoading={isLoading}
      />
    </div>
  );
};

interface GroupAdminComponentProps {
  groupId?: string;
  groupData?: GroupData;
}

const GroupAdminComponent: React.FC<GroupAdminComponentProps> = ({ groupId, groupData }) => {
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
      id: 'events',
      title: 'Управление событиями',
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
                <CurrentComponent groupData={groupData} groupId={groupId} />
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