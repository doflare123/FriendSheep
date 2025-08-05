'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

// Основной класс для работы с защитой страниц
export class PageProtection {
  /**
   * Предоставляет доступ к странице
   * @param pageKey - уникальный ключ страницы
   * @param expiry - время жизни доступа в секундах (по умолчанию 60 сек)
   */
  static grantAccess(pageKey: string, expiry: number = 60): void {
    const expiryTime = Date.now() + expiry * 1000;
    sessionStorage.setItem(`page_access_${pageKey}`, expiryTime.toString());
  }

  /**
   * Проверяет доступ к странице
   * @param pageKey - уникальный ключ страницы
   * @param autoRemove - автоматически удалить доступ после проверки (одноразовый доступ)
   */
  static checkAccess(pageKey: string, autoRemove: boolean = true): boolean {
    const key = `page_access_${pageKey}`;
    const expiryTime = sessionStorage.getItem(key);

    if (!expiryTime) {
      return false;
    }

    const now = Date.now();
    const expiry = parseInt(expiryTime, 10);

    if (now > expiry) {
      sessionStorage.removeItem(key);
      return false;
    }

    if (autoRemove) {
      sessionStorage.removeItem(key);
    }

    return true;
  }

  /**
   * Удаляет доступ к странице
   * @param pageKey - уникальный ключ страницы
   */
  static revokeAccess(pageKey: string): void {
    sessionStorage.removeItem(`page_access_${pageKey}`);
  }

  /**
   * Навигация с предоставлением доступа
   * @param router - Next.js router
   * @param path - путь для перехода
   * @param pageKey - ключ страницы
   * @param expiry - время жизни доступа в секундах
   */
  static navigateWithAccess(
    router: any,
    path: string, 
    pageKey: string, 
    expiry: number = 60
  ): void {
    this.grantAccess(pageKey, expiry);
    router.push(path);
  }
}

// Хук для защиты страниц
interface UsePageProtectionOptions {
  pageKey: string;
  redirectTo?: string;
  autoRemove?: boolean;
  onAccessDenied?: () => void;
}

export function usePageProtection({
  pageKey,
  redirectTo = '/',
  autoRemove = true,
  onAccessDenied
}: UsePageProtectionOptions) {
  const router = useRouter();

  useEffect(() => {
    const hasAccess = PageProtection.checkAccess(pageKey, autoRemove);

    if (!hasAccess) {
      if (onAccessDenied) {
        onAccessDenied();
      } else {
        router.push(redirectTo);
      }
    }
  }, [pageKey, redirectTo, autoRemove, router, onAccessDenied]);
}

// Готовые ключи для страниц (чтобы не ошибиться в названиях)
export const PAGE_KEYS = {
  REGISTRATION_COMPLETE: 'registration_complete',
  CODE_VERIFY: 'email_verified'
} as const;