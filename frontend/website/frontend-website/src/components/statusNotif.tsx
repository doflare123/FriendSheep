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

  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  const handleClose = useCallback((e?: React.MouseEvent) => {
    if (e) {
      e.stopPropagation();
    }
    setClosing(true);
    setTimeout(() => onCloseRef.current(), 300);
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => handleClose(), 3000);
    return () => clearTimeout(timer);
  }, [handleClose]);

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
      onClick={handleClose}
      style={{ cursor: 'pointer' }}
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

export default function StatusNotif() {
  const [notifs, setNotifs] = useState<NotifProps[]>([]);

  useEffect(() => {
    console.log("StatusNotif mounted, setting up listener");
    
    const handlePush = (e: Event) => {
      const event = e as CustomEvent;
      console.log("push-notif event received:", event.detail);
      
      const newNotif: NotifProps = {
        id: idCounter++,
        code: event.detail.code,
        description: event.detail.description,
        type: event.detail.type,
      };
      
      setNotifs((prev) => [...prev, newNotif]);
      console.log("Notification added to state");
    };

    window.addEventListener("push-notif", handlePush);
    console.log("Event listener registered");

    return () => {
      console.log("StatusNotif unmounting, removing listener");
      window.removeEventListener("push-notif", handlePush);
    };
  }, []);

  const removeNotif = useCallback((id: number) => {
    setNotifs((prev) => prev.filter((n) => n.id !== id));
  }, []);

  return (
    <div className={styles.container}>
      {notifs.map((notif) => (
        <NotifItem
          key={notif.id}
          {...notif}
          onClose={() => removeNotif(notif.id)}
        />
      ))}
    </div>
  );
}