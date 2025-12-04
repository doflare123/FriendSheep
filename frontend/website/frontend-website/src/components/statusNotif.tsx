// statusNotif.tsx
"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import Image from "next/image";
import styles from "../styles/statusNotif.module.css";

export type NotifProps = {
  id: number;
  code: number;
  description: string;
  type: "error" | "warning" | "success";
};

let idCounter = 0;

function NotifItem({
  id,
  code,
  description,
  type,
  onClose,
}: NotifProps & { onClose: () => void }) {
  const [closing, setClosing] = useState(false);
  const onCloseRef = useRef(onClose);

  // Обновляем ref при изменении onClose
  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  const handleClose = useCallback((e?: React.MouseEvent) => {
    if (e) {
      e.stopPropagation();
    }
    setClosing(true);
    setTimeout(() => onCloseRef.current(), 300);
  }, []); // Пустой массив зависимостей!

  // Автозакрытие через 3 секунды
  useEffect(() => {
    const timer = setTimeout(() => handleClose(), 3000);
    return () => clearTimeout(timer);
  }, []); // Пустой массив зависимостей!

  // Свайп
  useEffect(() => {
    let startX = 0;
    const el = document.getElementById(`notif-${id}`);
    if (!el) return;

    const handleDown = (e: MouseEvent) => {
      startX = e.clientX;
    };
    
    const handleUp = (e: MouseEvent) => {
      if (e.clientX - startX > 100) {
        handleClose();
      }
    };

    el.addEventListener("mousedown", handleDown);
    el.addEventListener("mouseup", handleUp);
    
    return () => {
      el.removeEventListener("mousedown", handleDown);
      el.removeEventListener("mouseup", handleUp);
    };
  }, [id, handleClose]);

  const iconMap = {
    error: "/notification/error.png",
    warning: "/notification/warning.png",
    success: "/notification/success.png",
  };

  const titleMap = {
    error: "Ошибка!",
    warning: "Осторожно!",
    success: "Успешно!",
  };

  const borderColor = {
    error: "var(--color-danger)",
    warning: "var(--color-warning)",
    success: "var(--color-success)",
  }[type];

  return (
    <div
      id={`notif-${id}`}
      className={`${styles.notif} ${closing ? styles.closing : ""}`}
    >
      <svg className={styles.progress} xmlns="http://www.w3.org/2000/svg">
        <rect
          x="1"
          y="1"
          rx="12"
          ry="12"
          width="calc(100% - 2px)"
          height="calc(100% - 2px)"
          stroke={borderColor}
          fill="none"
          pathLength="1"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
      
      <button className={styles.closeButton} onClick={handleClose}>
        <Image 
          src="/icons/close.png" 
          alt="Закрыть" 
          width={16} 
          height={16}
        />
      </button>

      <Image src={iconMap[type]} alt={type} width={24} height={24} />
      <div className={styles.content}>
        <strong className={styles.title}>{titleMap[type]}</strong>
        <p className={styles.desc}>{description}</p>
        <span className={styles.code}>Статус код: {code}</span>
      </div>
    </div>
  );
}