"use client";

import React, { useState } from 'react';
import Image from 'next/image';
import GroupCard from '../../components/Groups/GroupCard';
import GroupsScroll from '../../components/Groups/GroupsScroll';
import CreateGroupModal from '../../components/Groups/CreateGroupModal';
import '../../styles/Groups/GroupsPage.css';

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

  // Пример данных для групп под управлением
  const managedGroups: Group[] = [
    {
      id: '1',
      name: 'Мега крутая группа',
      description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к нам!',
      memberCount: 47,
      categories: ['games', 'movies'],
      socialLinks: {
        ds: 'discord.gg/example',
        tg: 't.me/example',
        vk: 'vk.com/example'
      },
      isPrivate: false,
      city: 'Москва'
    },
    {
      id: '2',
      name: 'Киноманы объединяйтесь',
      description: 'Обсуждаем фильмы, сериалы и всё что связано с кинематографом',
      memberCount: 123,
      categories: ['movies'],
      socialLinks: {
        tg: 't.me/cinephiles'
      },
      isPrivate: false,
      city: 'Санкт-Петербург'
    },
    {
      id: '3',
      name: 'Настольные игры СПб',
      description: 'Играем в настолки каждые выходные!',
      memberCount: 89,
      categories: ['board'],
      socialLinks: {
        vk: 'vk.com/boardgames_spb'
      },
      isPrivate: true,
      city: 'Санкт-Петербург'
    }
  ];

  // Пример данных для подписок
  const subscriptions: Group[] = [
    {
      id: '4',
      name: 'Геймеры Москвы',
      description: 'Крупнейшее сообщество геймеров столицы',
      memberCount: 1250,
      categories: ['games'],
      socialLinks: {
        ds: 'discord.gg/gamers_moscow',
        tg: 't.me/gamers_moscow'
      },
      isPrivate: false,
      city: 'Москва'
    },
    {
      id: '5',
      name: 'Книжный клуб',
      description: 'Читаем и обсуждаем интересные книги',
      memberCount: 67,
      categories: ['other'],
      socialLinks: {
        vk: 'vk.com/bookclub'
      },
      isPrivate: false,
      city: 'Москва'
    }
  ];

  const handleCreateGroup = (groupData: any) => {
    console.log('Creating group:', groupData);
    // Здесь будет логика создания группы
    setIsCreateModalOpen(false);
  };

  return (
    <div className='bgPage h-full'>
      <div className='contentContainer'>
        {/* Группы под управлением */}
        <div className='section'>
          <div className='sectionHeader'>
            <h2 className='sectionTitle'>Группы под вашим управлением</h2>
            <button 
              className='addButton'
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
          <GroupsScroll 
            groups={managedGroups} 
            emptyMessage="У вас еще нет групп под управлением"
            actionType="manage"
          />
        </div>

        {/* Подписки */}
        <div className='section'>
          <div className='sectionHeader'>
            <h2 className='sectionTitle'>Подписки</h2>
          </div>
          <GroupsScroll 
            groups={subscriptions}
            emptyMessage="У вас еще нет подписок"
            actionType="subscribe"
          />
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