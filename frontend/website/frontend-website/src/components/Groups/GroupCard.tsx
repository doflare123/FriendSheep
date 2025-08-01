import React from 'react';
import Image from 'next/image';
import '../../styles/Groups/GroupCard.css';

interface Group {
  id: number;
  name: string;
  small_description: string;
  member_count: number;
  image: string;
  category: string[];
  type: string;
}

interface GroupCardProps {
  group: Group;
  actionType: 'manage' | 'subscribe';
}

const GroupCard: React.FC<GroupCardProps> = ({ group, actionType }) => {
  const getCategoryIcon = (category: string) => {
    switch (category.toLowerCase()) {
      case 'игры': return '/events/games.png';
      case 'фильмы': return '/events/movies.png';
      case 'настольные игры': return '/events/board.png';
      default: return '/events/other.png';
    }
  };

  const getMemberCountText = (count: number) => {
    const lastDigit = count % 10;
    const lastTwoDigits = count % 100;

    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
      return `${count} участников`;
    }

    switch (lastDigit) {
      case 1:
        return `${count} участник`;
      case 2:
      case 3:
      case 4:
        return `${count} участника`;
      default:
        return `${count} участников`;
    }
  };

  const isPrivate = group.type === 'приватная группа';

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
            {group.category.map((categoryName, index) => (
              <div 
                key={index} 
                className='categoryIcon'
                title={categoryName}
              >
                <Image
                  src={getCategoryIcon(categoryName)}
                  alt={categoryName}
                  width={16}
                  height={16}
                />
              </div>
            ))}
          </div>

          <p className='groupMemberCount'>{getMemberCountText(group.member_count)}</p>
        </div>

        {/* Описание */}
        <div className='groupDescription'>
          {group.small_description}
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