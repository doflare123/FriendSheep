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
  socialLinks: {
    ds?: string;
    tg?: string;
    vk?: string;
  };
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

  const getSocialIcon = (social: string) => {
    return `/social/${social}.png`;
  };

  return (
    <div className='groupCard'>
      {/* Изображение группы */}
      <div className='groupImage'>
        <Image
          src={group.image || '/group-default.png'}
          alt={group.name}
          width={150}
          height={150}
          className='groupAvatar'
        />
        
        {/* Категории в правом верхнем углу */}
        <div className='categoryIcons'>
          {group.categories.map((category, index) => (
            <div key={index} className='categoryIcon'>
              <Image
                src={getCategoryIcon(category)}
                alt={category}
                width={16}
                height={16}
              />
            </div>
          ))}
        </div>

        {/* Индикатор приватности */}
        {group.isPrivate && (
          <div className='privateIcon'>
            🔒
          </div>
        )}
      </div>

      {/* Содержимое карточки */}
      <div className='groupContent'>
        <h3 className='groupName'>{group.name}</h3>
        <p className='groupMemberCount'>{group.memberCount} участников</p>
        
        <div className='groupDescription'>
          {group.description.split('\n').map((line, index) => (
            <React.Fragment key={index}>
              {line}
              {index < group.description.split('\n').length - 1 && <br />}
            </React.Fragment>
          ))}
        </div>

        {/* Социальные сети */}
        <div className='socialLinks'>
          {Object.entries(group.socialLinks).map(([social, link]) => (
            <div key={social} className='socialIcon'>
              <Image
                src={getSocialIcon(social)}
                alt={social}
                width={20}
                height={20}
              />
            </div>
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