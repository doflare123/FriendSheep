import React, { useState } from 'react';
import {EventCardProps} from "../../types/Events";
import '../../styles/EventCard.css';
import Image from "next/image";
import {getCategoryIcon} from '../../Constants';
import EventDetailModal from './EventDetailModal';

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
    onEdit
}) => {
    const [isModalOpen, setIsModalOpen] = useState(false);

    const handleButtonClick = () => {
        if (isEditMode && onEdit) {
            onEdit(id);
        } else {
            // Открываем модальное окно с подробностями события
            setIsModalOpen(true);
            console.log('Открываем подробности события:', id);
        }
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
    };

    const handleJoin = () => {
        console.log('Присоединиться к событию:', id);
        // Здесь будет логика присоединения к событию
    };

    return (
        <>
            <div className='eventCard'>
                <div className='cardImage'>
                    <img src={image} alt={title} />
                    <div className='typeIcon'>
                        <Image src={getCategoryIcon(type)} alt={type} width={20} height={20} />
                    </div>
                    <div className='cardDate'>{date}</div>
                </div>
                <div className='locationIcon'>
                    <Image src={`/events/${location}.png`} alt={location} width={20} height={20} />
                </div>
                <div className='cardContent'>
                    <h3 className='cardTitle'>{title}</h3>
                    <div className='cardGenres'>
                        <span className='genresLabel'>Жанры:</span>
                        {genres.map((genre, index) => (
                            <span key={index} className='genre'>{genre}</span>
                        ))}
                    </div>
                    <div className='cardMeta'>
                        <span className='cardParticipants'>
                            Участников: {participants}/{maxParticipants}
                            <Image src="/events/person.png" alt="person" width={16} height={16} className="personIcon" />
                        </span>
                        {duration && (
                            <span className='cardDuration'>
                                {duration}
                                <Image src="/events/clock.png" alt="clock" width={16} height={16} className="clockIcon" />
                            </span>
                        )}
                    </div>
                    <button className='detailsButton' onClick={handleButtonClick}>
                        {isEditMode ? 'Редактировать' : 'Узнать подробнее'}
                    </button>
                </div>
            </div>

            {/* Модальное окно с подробностями события */}
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