
export const formatTitle = (title: string): string => {
  const maxLength = 40;
  const firstLineLimit = 20;
  const secondLineLimit = 30;

  if (!title || title.trim().length < 5) {
    return "Без названия";
  }

  let trimmed = title.slice(0, maxLength);

  let firstLine = trimmed.slice(0, firstLineLimit);
  let secondLine = trimmed.slice(firstLineLimit, firstLineLimit + secondLineLimit);

  return firstLine + (secondLine ? "\n" + secondLine : "");
};
