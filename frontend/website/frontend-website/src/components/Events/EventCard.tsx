import {EventCardProps} from "../../types/Events"
import '../../styles/EventCard.css';
import Image from "next/image";

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
    location
}) => {
    const getTypeIcon = () => {
        switch (type) {
        case 'games': return '/events/games.png';
        case 'movies': return '/events/movies.png';
        case 'board': return '/events/board.png';
        default: return '/events/other.png';
        }
    };

    return (
        <div className='eventCard'>
            <div className='cardImage'>
                <img src={image} alt={title} />
                <div className='typeIcon'>
                    <Image src={getTypeIcon()} alt={type} width={20} height={20} />
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
                <button className='detailsButton'>
                    Узнать подробнее
                </button>
            </div>
        </div>
    );
};

export default EventCard;