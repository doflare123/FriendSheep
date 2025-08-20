import React from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import styles from '../../styles/Groups/GroupCard.module.css';
import {getCategoryIcon} from '../../Constants'

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
  const router = useRouter();

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

  const handleActionClick = () => {
    if (actionType === 'subscribe') {
      router.push(`/groups/profile/${group.id}`);
    } else {
      // Здесь можно добавить логику для управления группой
      router.push(`/groups/admin/${group.id}`);
    }
  };

  const isPrivate = group.type === 'приватная группа';

  return (
    <div className={styles.groupCard}>
      {/* Иконка группы в левом верхнем углу */}
      <div className={styles.groupImageContainer}>
        <Image
          src={group.image || '/default/group.jpg'}
          alt={group.name}
          width={80}
          height={80}
          className={styles.groupAvatar}
        />
      </div>

      {/* Основное содержимое справа от иконки */}
      <div className={styles.groupMainContent}>
        {/* Верхняя часть: название, категории, участники */}
        <div className={styles.groupTopSection}>
          <h3 className={styles.groupName} title={group.name}>
            {group.name}
          </h3>

          {/* Категории */}
          <div className={styles.categoryIcons}>
            {group.category.map((categoryName, index) => (
              <div key={index} className={styles.categoryIcon} title={categoryName}>
                <Image
                  src={getCategoryIcon(categoryName)}
                  alt={categoryName}
                  width={16}
                  height={16}
                />
              </div>
            ))}
          </div>

          <p className={styles.groupMemberCount}>
            {getMemberCountText(group.member_count)}
          </p>
        </div>

        {/* Описание */}
        <div className={styles.groupDescription}>
          {group.small_description}
        </div>

        {/* Кнопка действия */}
        <button className={styles.actionButton} onClick={handleActionClick}>
          {actionType === 'manage' ? 'Управлять' : 'Перейти'}
        </button>
      </div>
    </div>
  );
};

export default GroupCard;