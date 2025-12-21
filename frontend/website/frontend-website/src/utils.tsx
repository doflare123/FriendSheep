// notificationManager.ts
import { createRoot } from "react-dom/client";
import StatusNotif, { NotifProps } from "./components/statusNotif";

class NotificationManager {
  private container: HTMLDivElement | null = null;
  private root: ReturnType<typeof createRoot> | null = null;
  private initialized = false;
  private isInitializing = false;

  private async init() {
    if (this.initialized) {
      console.log("Manager already initialized");
      return;
    }
    
    if (this.isInitializing) {
      console.log("Manager is initializing, waiting...");
      // Ждем пока инициализация завершится
      while (this.isInitializing) {
        await new Promise(resolve => setTimeout(resolve, 10));
      }
      return;
    }
    
    console.log("Initializing manager");
    this.isInitializing = true;
    
    this.container = document.createElement("div");
    this.container.id = "notif-root";
    document.body.appendChild(this.container);
    this.root = createRoot(this.container);
    this.root.render(<StatusNotif />);
    
    // Даем время React на монтирование компонента и регистрацию слушателя
    await new Promise(resolve => setTimeout(resolve, 50));
    
    this.initialized = true;
    this.isInitializing = false;
    console.log("Manager initialized");
  }

  async show(code: number, description: string, type?: NotifProps["type"]) {
    console.log("showNotification called with:", code, description);
    await this.init();

    let resolvedType: NotifProps["type"];
    if (type) {
      resolvedType = type;
    } else if (code >= 200 && code < 300) {
      resolvedType = "success";
    } else if (code >= 300 && code < 400) {
      resolvedType = "warning";
    } else {
      resolvedType = "error";
    }

    console.log("Dispatching event");
    const event = new CustomEvent("push-notif", {
      detail: { code, description, type: resolvedType },
    });
    window.dispatchEvent(event);
  }
}

const manager = new NotificationManager();

export function showNotification(
  code: number,
  description: string,
  type?: NotifProps["type"]
) {
  console.log("=== showNotification wrapper called ===");
  manager.show(code, description, type);
}

export const getTheme = (): 'light' | 'dark' => {
  if (typeof window === 'undefined') return 'light';
  
  const saved = localStorage.getItem('theme');
  if (saved === 'dark' || saved === 'light') {
    return saved;
  }
  
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

export const setTheme = (theme: 'light' | 'dark') => {
  if (typeof window === 'undefined') return;
  
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
};

export const toggleTheme = () => {
  const current = getTheme();
  const newTheme = current === 'light' ? 'dark' : 'light';
  setTheme(newTheme);
  return newTheme;
};

export const initTheme = () => {
  if (typeof window === 'undefined') return;
  
  const theme = getTheme();
  document.documentElement.setAttribute('data-theme', theme);
};