import React, { useRef, useState, useEffect } from 'react';
import GroupCard from './GroupCard';
import '../../styles/Groups/GroupsScroll.css';

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

interface GroupsScrollProps {
  groups: Group[];
  emptyMessage: string;
  actionType: 'manage' | 'subscribe';
}

const GroupsScroll: React.FC<GroupsScrollProps> = ({ groups, emptyMessage, actionType }) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(true);

  const checkScrollButtons = () => {
    if (scrollRef.current) {
      const { scrollLeft, scrollWidth, clientWidth } = scrollRef.current;
      setCanScrollLeft(scrollLeft > 0);
      setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 1);
    }
  };

  useEffect(() => {
    const scrollElement = scrollRef.current;
    if (scrollElement && groups && groups.length > 0) {
      checkScrollButtons();
      scrollElement.addEventListener('scroll', checkScrollButtons);
      
      const resizeObserver = new ResizeObserver(checkScrollButtons);
      resizeObserver.observe(scrollElement);

      return () => {
        scrollElement.removeEventListener('scroll', checkScrollButtons);
        resizeObserver.disconnect();
      };
    }
  }, [groups]);

  const scroll = (direction: 'left' | 'right') => {
    if (scrollRef.current) {
      const scrollAmount = 320; // ширина карточки группы + отступы
      scrollRef.current.scrollBy({
        left: direction === 'left' ? -scrollAmount : scrollAmount,
        behavior: 'smooth'
      });
    }
  };

  // Проверяем и null, и пустой массив
  if (!groups || groups.length === 0) {
    return (
      <div className='emptyGroupsMessage'>
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className='groupsContainer'>
      {canScrollLeft && (
        <button 
          className='scrollButton scrollLeft'
          onClick={() => scroll('left')}
        >
          ←
        </button>
      )}
      
      <div className='groupsScroll' ref={scrollRef}>
        {groups.map((group) => (
          <GroupCard 
            key={group.id} 
            group={group} 
            actionType={actionType}
          />
        ))}
      </div>
      
      {canScrollRight && (
        <button 
          className='scrollButton scrollRight'
          onClick={() => scroll('right')}
        >
          →
        </button>
      )}
    </div>
  );
};

export default GroupsScroll;