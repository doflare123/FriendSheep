"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import styles from "../styles/statusNotif.module.css";

export type NotifProps = {
  id: number;
  code: number;
  description: string;
  type: "error" | "warning" | "success";
};

let idCounter = 0;

export default function StatusNotif() {
  const [notifs, setNotifs] = useState<NotifProps[]>([]);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent).detail as Omit<NotifProps, "id">;
      const newNotif = { ...detail, id: ++idCounter };
      setNotifs((prev) => [...prev, newNotif]);
    };
    window.addEventListener("push-notif", handler);
    return () => window.removeEventListener("push-notif", handler);
  }, []);

  const removeNotif = (id: number) => {
    setNotifs((prev) => prev.filter((n) => n.id !== id));
  };

  return (
    <div className={styles.container}>
      {notifs.map((notif) => (
        <NotifItem key={notif.id} {...notif} onClose={() => removeNotif(notif.id)} />
      ))}
    </div>
  );
}

function NotifItem({
  id,
  code,
  description,
  type,
  onClose,
}: NotifProps & { onClose: () => void }) {
  const [closing, setClosing] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => handleClose(), 15000);
    return () => clearTimeout(timer);
  }, []);

  const handleClose = () => {
    setClosing(true);
    setTimeout(onClose, 300);
  };

  // свайп
  useEffect(() => {
    let startX = 0;
    const handleDown = (e: MouseEvent) => (startX = e.clientX);
    const handleUp = (e: MouseEvent) => {
      if (e.clientX - startX > 100) handleClose();
    };
    const el = document.getElementById(`notif-${id}`);
    el?.addEventListener("mousedown", handleDown);
    el?.addEventListener("mouseup", handleUp);
    return () => {
      el?.removeEventListener("mousedown", handleDown);
      el?.removeEventListener("mouseup", handleUp);
    };
  }, []);

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
    error: "#E6374B",
    warning: "#FF6A00",
    success: "#51A732",
  }[type];

    return (
        <div
            id={`notif-${id}`}
            className={`${styles.notif} ${closing ? styles.closing : ""}`}
            onClick={handleClose}
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
            <Image src={iconMap[type]} alt={type} width={24} height={24} />
            <div className={styles.content}>
                <strong className={styles.title}>{titleMap[type]}</strong>
                <p className={styles.desc}>{description}</p>
                <span className={styles.code}>Статус код: {code}</span>
            </div>
        </div>
    );
}

