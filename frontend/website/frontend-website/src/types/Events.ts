export interface EventCardProps {
    id: number;
    type: 'games' | 'movies' | 'board' | 'other';
    image: string;
    date: string;
    title: string;
    genres: string[];
    participants: number;
    maxParticipants: number;
    duration?: string;
    location: string;
    scale?: number;
}

export interface SectionData {
    title: string;
    pattern: string;
    categories: EventCardProps[];
}