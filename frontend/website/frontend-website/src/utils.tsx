import { createRoot } from "react-dom/client";
import StatusNotif, { NotifProps } from "./components/statusNotif";

let container: HTMLDivElement | null = null;
let root: ReturnType<typeof createRoot> | null = null;

export function showNotification(
  code: number,
  description: string,
  type?: NotifProps["type"]
) {
  console.log("ZZZ", code, description)
  if (!container) {
    container = document.createElement("div");
    container.id = "notif-root";
    document.body.appendChild(container);
    root = createRoot(container);
  }

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

  const event = new CustomEvent("push-notif", {
    detail: { code, description, type: resolvedType },
  });
  window.dispatchEvent(event);

  // отрисовываем root контейнер
  root?.render(<StatusNotif />);
}
