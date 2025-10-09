// src/types/Events.ts

export interface EventCardProps {
    id: number;
    type: 'games' | 'movies' | 'board' | 'other'; 
    image: string;
    date: string;                     // серверный start_time
    start_time?: string;
    end_time?: string;                // 🔥 добавлено (для будущих эвентов)
    title: string;
    genres: string[];
    participants: number;             // серверный current_users
    maxParticipants: number;          // серверный max_users
    duration?: string;
    location: 'online' | 'offline';   // пока заглушка, сервер не даёт
    adress: string;                   // серверный location
    publisher?: string;
    city?: string;
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
