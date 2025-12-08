import React, { useState } from 'react';
import {EventCardProps} from "../../types/Events";
import styles from '../../styles/EventCard.module.css';
import Image from "next/image";
import {getCategoryIcon, convertMinutesToReadableTime} from '../../Constants';
import EventDetailModal from './EventDetailModal';
import { showNotification } from "@/utils";

const EventCard: React.FC<EventCardProps> = ({
    id,
    type,
    image,
    date,
    title,
    genres,
    participants,
    maxParticipants,
    duration,
    location,
    isEditMode = false,
    city,
    onEdit
}) => {
    const [isModalOpen, setIsModalOpen] = useState(false);

    const handleButtonClick = () => {
        if (isEditMode && onEdit) {
            onEdit(id);
        } else {
            setIsModalOpen(true);
            console.log('Открываем подробности события:', id);
        }
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
    };

    const handleJoin = () => {
        console.log('Присоединиться к событию:', id);
    };

    console.log("Title", title);

    return (
        <>
            <div className={styles.eventCard}>
                <div className={styles.cardImage}>
                    <img src={image || "/default/event_card.jpg"} alt={title} />
                    <div className={styles.typeIcon}>
                        <Image src={getCategoryIcon(type)} alt={type} width={20} height={20} />
                    </div>
                    <div className={styles.cardDate} title="По местному времени">{date}</div>

                    {location === 'offline' && (
                        <div className={styles.cityBadge}>
                            <span>{city || 'Калининград'}</span>
                        </div>
                    )}
                </div>
                <div className={styles.locationIcon}>
                    <Image src={`/events/${location}.png`} alt={location} width={20} height={20} />
                </div>
                <div className={styles.cardContent}>
                    <h3 className={styles.cardTitle}>{title}</h3>
                    <div className={styles.cardGenres}>
                        <span className={styles.genresLabel}>Жанры:</span>
                        {genres.map((genre, index) => (
                            <span key={index} className={styles.genre}>{genre}</span>
                        ))}
                    </div>
                    <div className={styles.cardMeta}>
                        <span className={styles.cardParticipants}>
                            Участников: {participants}/{maxParticipants}
                            <Image src="/events/person.png" alt="person" width={16} height={16} className={styles.personIcon} />
                        </span>
                        {duration && (
                            <span className={styles.cardDuration}>
                                {convertMinutesToReadableTime(duration)}
                                <Image src="/events/clock.png" alt="clock" width={16} height={16} className={styles.clockIcon} />
                            </span>
                        )}
                    </div>
                    <button className={styles.detailsButton} onClick={handleButtonClick}>
                        {isEditMode ? 'Редактировать' : 'Узнать подробнее'}
                    </button>
                </div>
            </div>

            <EventDetailModal
                isOpen={isModalOpen}
                onClose={handleCloseModal}
                onJoin={handleJoin}
                eventId={id}
            />
        </>
    );
};

export default EventCard;