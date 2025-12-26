import React, { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import styles from '@/styles/Events/EventDetailModal.module.css';
import { getEventInfo } from '@/api/events/getEventInfo';
import { joinEvent } from '@/api/events/joinEvent';
import { leaveEvent } from '@/api/events/leaveEvent';
import { getGroupInfo } from '@/api/get_group_info';
import { getOtherUserInfo } from '@/api/profile/getProfile';
import { getUserInfo } from '@/api/profile/getOwnProfile';
import { getAccesToken, convertSingleCategRuToEng, convertSessionPlaceToLocation, getUserData } from '@/Constants';
import { showNotification } from '@/utils';
import { getOwnGroups } from '@/api/get_owngroups';
import LoadingIndicator from '@/components/LoadingIndicator';
import VerifiedBadge from '@/components/VerifiedBadge';
import type { EventFullResponse } from '@/types/apiTypes';

interface EventDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onParticipantsChange?: (newCount: number) => void;
  eventId?: number;
}

const EventDetailModal: React.FC<EventDetailModalProps> = ({ 
  isOpen, 
  onClose, 
  onParticipantsChange,
  eventId 
}) => {
  const [eventData, setEventData] = useState<EventFullResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [isCreator, setIsCreator] = useState(false);
  const [groupName, setGroupName] = useState<string>('');
  const [groupId, setGroupId] = useState<number | null>(null);
  const [isVerifiedCreator, setIsVerifiedCreator] = useState(false);
  const router = useRouter();

  useEffect(() => {
    if (isOpen && eventId) {
      loadEventData();
    }
    
    if (!isOpen) {
      setEventData(null);
      setIsCreator(false);
      setGroupName('');
      setGroupId(null);
      setIsVerifiedCreator(false);
    }
  }, [isOpen, eventId]);

  const checkIfCreator = async (groupId: number, accessToken: string) => {
    try {
      const ownGroups = await getOwnGroups(accessToken);
      const isOwner = ownGroups.some((group: any) => group.id === groupId);
      setIsCreator(isOwner);
    } catch (error) {
      console.error('Ошибка при проверке создателя:', error);
      setIsCreator(false);
    }
  };

  const loadGroupInfo = async (groupId: number, accessToken: string) => {
    try {
      const groupData = await getGroupInfo(accessToken, groupId);
      setGroupName(groupData.name);
      setGroupId(groupId);

      if (groupData.creater) {
        try {
          const currentUser = getUserData();
          let creatorData;

          if (currentUser && currentUser.id === groupData.creater) {
            creatorData = await getUserInfo(accessToken);
          } else {
            creatorData = await getOtherUserInfo(accessToken, groupData.creater);
          }

          setIsVerifiedCreator(creatorData.enterprise || false);
        } catch (error) {
          console.error('Ошибка при получении данных создателя:', error);
          setIsVerifiedCreator(false);
        }
      }
    } catch (error) {
      console.error('Ошибка при загрузке информации о группе:', error);
    }
  };

  const loadEventData = async () => {
    if (!eventId) return;

    setIsLoading(true);
    try {
      const accessToken = await getAccesToken(router);
      const response = await getEventInfo(accessToken, eventId);
      setEventData(response as EventFullResponse);
      
      await checkIfCreator(response.session.group_id, accessToken);
      await loadGroupInfo(response.session.group_id, accessToken);
    } catch (error: any) {
      console.error('Ошибка загрузки события:', error);
      const errorMessage = error.response?.data?.message || 'Не удалось загрузить данные события';
      const errorCode = error.response?.status || 500;
      showNotification(errorCode, errorMessage);
      onClose();
    } finally {
      setIsLoading(false);
    }
  };

  const handleJoinOrLeave = async () => {
    if (!eventData || isActionLoading || isCreator) return;

    setIsActionLoading(true);
    try {
      const accessToken = await getAccesToken(router);
      const { session } = eventData;

      if (session.is_sub) {
        await leaveEvent(accessToken, session.id);
        showNotification(200, 'Вы успешно отписались от события');
        
        if (onParticipantsChange) {
          onParticipantsChange(session.current_users - 1);
        }
      } else {
        await joinEvent(accessToken, session.group_id, session.id);
        showNotification(200, 'Вы успешно присоединились к событию');
        
        if (onParticipantsChange) {
          onParticipantsChange(session.current_users + 1);
        }
      }

      await loadEventData();
    } catch (error: any) {
      console.error('Ошибка при присоединении/отписке:', error);
      const errorMessage = error.response?.data?.message || 'Не удалось выполнить действие';
      const errorCode = error.response?.status || 500;
      showNotification(errorCode, errorMessage);
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleGroupClick = () => {
    if (groupId) {
      router.push(`/groups/profile/${groupId}`);
    }
  };

  if (!isOpen) return null;

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const year = date.getFullYear();
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    
    return `${day}.${month}.${year} ${hours}:${minutes}`;
  };

  const formatDuration = (minutes: number) => {
    return `${minutes} минут`;
  };

  const isUrl = (string: string) => {
    try {
      new URL(string);
      return true;
    } catch {
      return false;
    }
  };

  const getPublisherLink = (publisher?: string) => {
    if (!publisher) return null;
    return `https://yandex.ru/search/?text=${encodeURIComponent(publisher)}`;
  };

  const getLocationLink = (location: string) => {
    if (isUrl(location)) {
      return location;
    }
    return `https://yandex.ru/search/?text=${encodeURIComponent(location)}`;
  };

  if (isLoading || !eventData) {
    const modalContent = (
      <div className={styles.modalOverlay} onClick={onClose}>
        <div className={styles.modalContainer} onClick={(e) => e.stopPropagation()}>
          <LoadingIndicator text="Загрузка события..." />
        </div>
      </div>
    );
    return createPortal(modalContent, document.body);
  }

  const { session, metadata } = eventData;
  const displayGenres = metadata.Genres.slice(0, 9);
  const eventType = convertSingleCategRuToEng(session.session_type);
  const locationIcon = convertSessionPlaceToLocation(session.session_place);

  const modalContent = (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modalContainer} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <h2 className={styles.modalTitle}>{session.title}</h2>
          <button className={styles.closeButton} onClick={onClose}>
            <Image 
              src="/icons/close.png" 
              alt="Закрыть" 
              width={24} 
              height={24}
            />
          </button>
        </div>

        <div className={styles.eventImageContainer}>
          <Image
            src={session.image_url}
            alt={session.title}
            fill
            className={styles.eventImage}
          />
          <div className={styles.participantsBadge}>
            <Image 
              src={`/events/${locationIcon}.png`} 
              alt={locationIcon} 
              width={20} 
              height={20}
            />
          </div>
        </div>

        <div className={styles.eventInfo}>
          <div className={styles.dateInfo}>
            <span className={styles.dateValue} title="По местному времени">Дата проведения: {formatDate(session.start_time)}</span>
          </div>
          <div className={styles.durationInfo}>
            <span>{formatDuration(session.duration)}</span>
            <Image 
              src="/events/clock.png" 
              alt="Длительность" 
              width={16} 
              height={16}
            />
          </div>
        </div>

        <div className={styles.eventDescription}>
          <p>{metadata.Notes}</p>
        </div>

        {displayGenres.length > 0 && (
          <div className={styles.genresSection}>
            <span className={styles.genresLabel}>Жанры:</span>
            <div className={styles.genresContainer}>
              {displayGenres.map((genre, index) => (
                <span key={index} className={styles.genreTag}>
                  {genre}
                </span>
              ))}
            </div>
          </div>
        )}

        {(groupName || metadata.Fields?.publisher || metadata.Year || metadata.Location || metadata.AgeLimit) && (
          <div className={styles.publisherInfo}>
            {groupName && (
              <div className={styles.publisherRow}>
                <span className={styles.publisherLabel}>Группа:</span>
                <div 
                  className={`${styles.publisherValue} ${styles.underlined} ${styles.groupName}`}
                  onClick={handleGroupClick}
                >
                  {groupName}
                  <VerifiedBadge isVerified={isVerifiedCreator} size={16} />
                </div>
              </div>
            )}
            {metadata.Fields?.publisher && (
              <div className={styles.publisherRow}>
                <span className={styles.publisherLabel}>Издатель:</span>
                <a 
                  href={getPublisherLink(metadata.Fields.publisher) || '#'}
                  target="_blank" 
                  rel="noopener noreferrer"
                  className={`${styles.publisherValue} ${styles.underlined}`}
                >
                  {metadata.Fields.publisher}
                </a>
              </div>
            )}
            {metadata.Year && (
              <div className={styles.publisherRow}>
                <span className={styles.publisherLabel}>Год издания:</span>
                <span className={styles.publisherValue}>{metadata.Year}</span>
              </div>
            )}
            {metadata.Location && (
              <div className={styles.publisherRow}>
                <span className={styles.publisherLabel}>Место проведения:</span>
                <a 
                  href={getLocationLink(metadata.Location)}
                  target="_blank" 
                  rel="noopener noreferrer"
                  className={`${styles.publisherValue} ${styles.underlined}`}
                >
                  {metadata.Location}
                </a>
              </div>
            )}
            {metadata.AgeLimit && (
              <div className={styles.publisherRow}>
                <span className={styles.publisherLabel}>Возрастное ограничение:</span>
                <span className={styles.publisherValue}>{metadata.AgeLimit}</span>
              </div>
            )}
          </div>
        )}

        {isCreator ? (
          <div className={`${styles.joinButton} ${styles.creatorButton}`}>
            Вы создатель
          </div>
        ) : (
          <button 
            className={`${styles.joinButton} ${session.is_sub ? styles.leaveButton : ''}`}
            onClick={handleJoinOrLeave}
            disabled={isActionLoading}
          >
            {isActionLoading ? 'Загрузка...' : (session.is_sub ? 'Отписаться' : 'Присоединиться')}
          </button>
        )}
      </div>
    </div>
  );

  return createPortal(modalContent, document.body);
};

export default EventDetailModal;