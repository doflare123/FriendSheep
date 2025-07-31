import React from 'react';
import Image from 'next/image';
import '../../styles/Groups/GroupCard.css';

interface Group {
  id: string;
  name: string;
  description: string;
  memberCount: number;
  image?: string;
  categories: ('games' | 'movies' | 'board' | 'other')[];
  isPrivate: boolean;
  city: string;
}

interface GroupCardProps {
  group: Group;
  actionType: 'manage' | 'subscribe';
}

const GroupCard: React.FC<GroupCardProps> = ({ group, actionType }) => {
  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'games': return '/events/games.png';
      case 'movies': return '/events/movies.png';
      case 'board': return '/events/board.png';
      default: return '/events/other.png';
    }
  };

  const getCategoryName = (category: string) => {
    switch (category) {
      case 'games': return 'Игры';
      case 'movies': return 'Фильмы';
      case 'board': return 'Настольные игры';
      default: return 'Другое';
    }
  };

  return (
    <div className='groupCard'>
      {/* Иконка группы в левом верхнем углу */}
      <div className='groupImageContainer'>
        <Image
          src={group.image || '/default/group.jpg'}
          alt={group.name}
          width={80}
          height={80}
          className='groupAvatar'
        />
      </div>

      {/* Основное содержимое справа от иконки */}
      <div className='groupMainContent'>
        {/* Верхняя часть: название, категории, участники */}
        <div className='groupTopSection'>
          <h3 className='groupName' title={group.name}>{group.name}</h3>
          
          {/* Категории */}
          <div className='categoryIcons'>
            {group.categories.map((category, index) => (
              <div 
                key={index} 
                className='categoryIcon'
                title={getCategoryName(category)}
              >
                <Image
                  src={getCategoryIcon(category)}
                  alt={category}
                  width={16}
                  height={16}
                />
              </div>
            ))}
          </div>

          <p className='groupMemberCount'>{group.memberCount} участников</p>
        </div>

        {/* Описание */}
        <div className='groupDescription'>
          {group.description.split('\n').map((line, index) => (
            <React.Fragment key={index}>
              {line}
              {index < group.description.split('\n').length - 1 && <br />}
            </React.Fragment>
          ))}
        </div>

        {/* Кнопка действия */}
        <button className='actionButton'>
          {actionType === 'manage' ? 'Управлять' : 'Перейти'}
        </button>
      </div>
    </div>
  );
};

export default GroupCard;