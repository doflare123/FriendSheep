// notificationManager.ts
import { createRoot } from "react-dom/client";
import StatusNotif, { NotifProps } from "./components/statusNotif";

class NotificationManager {
  private container: HTMLDivElement | null = null;
  private root: ReturnType<typeof createRoot> | null = null;
  private initialized = false;

  private init() {
    if (this.initialized) {
      console.log("Manager already initialized");
      return;
    }
    
    console.log("Initializing manager");
    this.container = document.createElement("div");
    this.container.id = "notif-root";
    document.body.appendChild(this.container);
    this.root = createRoot(this.container);
    this.root.render(<StatusNotif />);
    this.initialized = true;
  }

  show(code: number, description: string, type?: NotifProps["type"]) {
    console.log("showNotification called with:", code, description);
    this.init();

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