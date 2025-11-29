"use client";

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import GroupsScroll from '../../components/Groups/GroupsScroll';
import CreateGroupModal from '../../components/Groups/CreateGroupModal';
import styles from '../../styles/Groups/GroupsPage.module.css';
import { createGroup } from '../../api/add_group';
import { getGroups } from '../../api/get_groups';
import { getOwnGroups } from '../../api/get_owngroups';
import { convertCategoriesToIds, convertSocialContactsToString, getAccesToken } from '../../Constants';
import {SmallGroup} from '@/types/Groups';

export default function GroupsPage() {
  const router = useRouter();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [managedGroups, setManagedGroups] = useState<SmallGroup[]>([]);
  const [subscriptions, setSubscriptions] = useState<SmallGroup[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Загрузка групп пользователя
  const loadUserGroups = async () => {
    try {
      setIsLoading(true);

      const accessToken = await getAccesToken(router);
      
      if (!accessToken) {
        //router.push('/login');
        return;
      }

      const response = await getGroups(accessToken);
      
      setSubscriptions(response);

      const responseOwn = await getOwnGroups(accessToken);
      

      setManagedGroups(responseOwn);
      
    } catch (error) {
      console.error('Ошибка при загрузке групп:', error);
      // Если ошибка связана с авторизацией (например, 401), перенаправляем на регистрацию
      if (error && typeof error === 'object' && 'status' in error && error.status === 401) {
        localStorage.removeItem('access_token');
        router.push('/register');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateGroup = async (groupData: any) => {
    setIsCreateModalOpen(false);

    try {
      const accessToken = await getAccesToken(router);
      if (!accessToken) {
        //router.push('/register');
        return;
      }

      const categories = convertCategoriesToIds(groupData.categories);
      const socialContacts = convertSocialContactsToString(groupData.socialContacts);

      await createGroup(
        groupData.name, 
        groupData.description, 
        groupData.shortDescription, 
        groupData.city, 
        categories, 
        groupData.isPrivate, 
        groupData.image, 
        socialContacts, 
        accessToken
      );

      // Перезагружаем группы после создания новой
      await loadUserGroups();
    } catch (error) {
      console.error('Ошибка при создании группы:', error);
      // Если ошибка связана с авторизацией, перенаправляем на регистрацию
      if (error && typeof error === 'object' && 'status' in error && error.status === 401) {
        localStorage.removeItem('access_token');
        router.push('/register');
      }
    }
  };

  useEffect(() => {
    loadUserGroups();
  }, []);

  return (
    <div className='bgPage' style={{ display: 'flex' }}>
      <div className={styles.contentContainer}>
        {/* Группы под управлением */}
        <div className={styles.section}>
          <div className={styles.sectionHeader}>
            <h2 className={styles.sectionTitle}>Группы под вашим управлением</h2>
            <button 
              className={styles.addButton}
              onClick={() => setIsCreateModalOpen(true)}
            >
              <Image
                src="/add_button.png"
                alt="Add group"
                width={24}
                height={24}
              />
            </button>
          </div>
          {isLoading ? (
            <div>Загрузка...</div>
          ) : (
            <GroupsScroll 
              groups={managedGroups} 
              emptyMessage="У вас еще нет групп под управлением"
              actionType="manage"
            />
          )}
        </div>

        {/* Подписки */}
        <div className={styles.section}>
          <div className={styles.sectionHeader}>
            <h2 className={styles.sectionTitle}>Подписки</h2>
          </div>
          {isLoading ? (
            <div>Загрузка...</div>
          ) : (
            <GroupsScroll 
              groups={subscriptions}
              emptyMessage="У вас еще нет подписок"
              actionType="subscribe"
            />
          )}
        </div>
      </div>

      {/* Модальное окно создания группы */}
      {isCreateModalOpen && (
        <CreateGroupModal
          onClose={() => setIsCreateModalOpen(false)}
          onSubmit={handleCreateGroup}
        />
      )}
    </div>
  );
}