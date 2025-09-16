// pages/index.tsx (или app/page.tsx для App Router)
import Image from "next/image";
import Footer from "../components/Footer";
import styles from '../styles/MainPage.module.css';
import { SectionData } from "../types/Events"
import CategorySection from '../components/Events/CategorySection'

export default function Home() {
  // Пример данных

  const mainSections: SectionData[] = [
    {
      title: "Популярные события",
      pattern: "/events/popular_bg.png",
      categories: [
        {
          id: "1",
          type: "games",
          image: "/event_card.jpg",
          date: "25 июля",
          title: "Командный шутер",
          genres: ["Шутер", "Многопользовательская"],
          participants: 5,
          maxParticipants: 8,
          location: "online"
        },
        {
          id: "2",
          type: "games",
          image: "/event_card.jpg",
          date: "26 июля",
          title: "Стратегия в реальном времени",
          genres: ["Стратегия", "RTS"],
          participants: 3,
          maxParticipants: 6,
          duration: "2 часа",
          location: "online"
        },
        {
          id: "3",
          type: "games",
          image: "/event_card.jpg",
          date: "27 июля",
          title: "Исследование подземелий",
          genres: ["Песочница", "Выживание"],
          participants: 4,
          maxParticipants: 6,
          duration: "3 часа",
          location: "online"
        },
        {
          id: "4",
          type: "movies",
          image: "/event_card.jpg",
          date: "28 июля",
          title: "Фантастический триллер",
          genres: ["Фантастика", "Триллер"],
          participants: 12,
          maxParticipants: 20,
          duration: "2ч 15мин",
          location: "online"
        }
      ]
    },
    {
      title: "Новые события",
      pattern: "/events/new_bg.png",
      categories: [
        {
          id: "5",
          type: "games",
          image: "/event_card.jpg",
          date: "30 июля",
          title: "Новая MMORPG",
          genres: ["MMORPG", "Фэнтези"],
          participants: 2,
          maxParticipants: 10,
          location: "online"
        },
        {
          id: "6",
          type: "movies",
          image: "/event_card.jpg",
          date: "1 августа",
          title: "Премьера комедии",
          genres: ["Комедия", "Семейный"],
          participants: 8,
          maxParticipants: 15,
          duration: "1ч 45мин",
          location: "online"
        }
      ]
    }
  ];

  const additionalSections: SectionData[] = [
    {
      title: "Фильмы",
      pattern: "/events/new_bg.png",
      categories: [
        {
          id: "7",
          type: "movies",
          image: "/event_card.jpg",
          date: "2 августа",
          title: "Классика кинематографа",
          genres: ["Драма", "Классика"],
          participants: 6,
          maxParticipants: 12,
          duration: "2ч 30мин",
          location: "online"
        },
        {
          id: "8",
          type: "movies",
          image: "/event_card.jpg",
          date: "3 августа",
          title: "Аниме-марафон",
          genres: ["Аниме", "Приключения"],
          participants: 15,
          maxParticipants: 25,
          duration: "4 часа",
          location: "online"
        }
      ]
    },
    {
      title: "Игры",
      pattern: "/events/game_bg.png",
      categories: [
        {
          id: "9",
          type: "games",
          image: "/event_card.jpg",
          date: "4 августа",
          title: "Турнир по Dota 2",
          genres: ["MOBA", "Соревновательная"],
          participants: 10,
          maxParticipants: 10,
          duration: "1 час",
          location: "online"
        }
      ]
    },
    {
      title: "Настольные игры",
      pattern: "/events/board_bg.png",
      categories: [
        {
          id: "10",
          type: "board",
          image: "/event_card.jpg",
          date: "5 августа",
          title: "Классическая мафия",
          genres: ["Логическая", "Ролевая"],
          participants: 8,
          maxParticipants: 12,
          duration: "1ч 30мин",
          location: "online"
        },
        {
          id: "11",
          type: "board",
          image: "/event_card.jpg",
          date: "6 августа",
          title: "Монополия: Московская версия",
          genres: ["Экономическая", "Семейная"],
          participants: 4,
          maxParticipants: 6,
          duration: "2 часа",
          location: "online"
        }
      ]
    },
    {
      title: "Другое",
      pattern: "/events/other_bg.png",
      categories: [
        {
          id: "12",
          type: "other",
          image: "/event_card.jpg",
          date: "7 августа",
          title: "Пешая экскурсия по центру",
          genres: ["Туризм", "История"],
          participants: 12,
          maxParticipants: 20,
          duration: "3 часа",
          location: "online"
        },
        {
          id: "13",
          type: "other",
          image: "/event_card.jpg",
          date: "8 августа",
          title: "Изучение японских иероглифов",
          genres: ["Обучение", "Языки"],
          participants: 5,
          maxParticipants: 8,
          duration: "1ч 30мин",
          location: "offline"
        }
      ]
    }
  ];

  return (
    <div>
      <div className='bgPage'>
        {/* Главные категории */}
        {mainSections.map((section, sectionIndex) => (
          <div key={sectionIndex} className={styles.section}>
            <CategorySection section={section} title={section.title} clickable={section.title!="Новые события"}/>
          </div>
        ))}

        {/* Заголовок "Категории" */}
        <div className={styles.categoriesHeader}>
          <h2>Категории</h2>
        </div>

        {/* Дополнительные категории */}
        {additionalSections.map((section, sectionIndex) => (
          <div key={sectionIndex} className={styles.section}>
            <CategorySection section={section} title={section.title} />
          </div>
        ))}
      </div>
      <Footer/>
    </div>
  );
}