"use client";

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import GroupsScroll from '../../components/Groups/GroupsScroll';
import CreateGroupModal from '../../components/Groups/CreateGroupModal';
import '../../styles/Groups/GroupsPage.css';
import { createGroup } from '../../api/add_group';
import { getGroups } from '../../api/get_groups';
import { getOwnGroups } from '../../api/get_owngroups';
import { convertCategoriesToIds, convertSocialContactsToString, convertIdsToCategories } from '../../Constants';

interface Group {
  id: string;
  name: string;
  description: string;
  memberCount: number;
  image?: string;
  categories: ('games' | 'movies' | 'board' | 'other')[];
  socialLinks: {
    ds?: string;
    tg?: string;
    vk?: string;
  };
  isPrivate: boolean;
  city: string;
}

export default function GroupsPage() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [managedGroups, setManagedGroups] = useState<Group[]>([]);
  const [subscriptions, setSubscriptions] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Загрузка групп пользователя
  const loadUserGroups = async () => {
    try {
      setIsLoading(true);
      const accessToken = localStorage.getItem('access_token');
      
      if (!accessToken) {
        console.error('Access token not found');
        return;
      }

      const response = await getGroups(accessToken);
      
      console.log("transformedGroups", response);
      setSubscriptions(response);

      const responseOwn = await getOwnGroups(accessToken);
      
      console.log("transformedGroupsOwn", responseOwn);
      setManagedGroups(responseOwn);
      
    } catch (error) {
      console.error('Ошибка при загрузке групп:', error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadUserGroups();
  }, []);

  const handleCreateGroup = async (groupData: any) => {
    console.log('Creating group:', groupData);
    setIsCreateModalOpen(false);

    try {
      const accessToken = localStorage.getItem('access_token');
      if (!accessToken) {
        console.error('Access token not found');
        return;
      }

      const categories = convertCategoriesToIds(groupData.categories);
      const socialContacts = convertSocialContactsToString(groupData.socialContacts);

      console.log("groupData.image", groupData.image);

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
    }
  };

  return (
    <div className="bgPage h-full">
      <div className="contentContainer">
        {/* Группы под управлением */}
        <div className="section">
          <div className="sectionHeader">
            <h2 className="sectionTitle">Группы под вашим управлением</h2>
            <button 
              className="addButton"
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
        <div className="section">
          <div className="sectionHeader">
            <h2 className="sectionTitle">Подписки</h2>
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