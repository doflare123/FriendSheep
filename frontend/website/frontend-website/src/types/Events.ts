export interface EventCardProps {
    id: string;
    type: 'games' | 'movies' | 'board' | 'other';
    image: string;
    date: string;
    title: string;
    genres: string[];
    participants: number;
    maxParticipants: number;
    duration?: string;
    location: string;
}

export interface SectionData {
    title: string;
    pattern: string;
    categories: EventCardProps[];
}