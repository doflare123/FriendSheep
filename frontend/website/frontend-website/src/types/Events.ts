// src/types/Events.ts

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
    location: 'online' | 'offline'; // Изменено: теперь содержит только статус онлайн/оффлайн
    adress: string; // Изменено: содержит фактический адрес или ссылку
    scale?: number;
    isEditMode?: boolean;
    onEdit?: (id: number) => void;
    groupId?: number;
}

export interface SectionData {
    title: string;
    pattern: string;
    categories: EventCardProps[];
}