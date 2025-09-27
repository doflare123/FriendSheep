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
    location: 'online' | 'offline'; // онлайн/оффлайн
    adress: string; // фактический адрес или ссылка
    city?: string;  // ✅ новое свойство
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
