import { SearchUserItem } from '@/api/types/user';
import { User } from '@/components/profile/UserCard';

export const createUserWithHighlightedText = (
  user: SearchUserItem,
  query: string
): User => {
  const trimmedQuery = query.trim();

  const isUsernameSearch = trimmedQuery.startsWith('@');
  const searchQuery = isUsernameSearch ? trimmedQuery.substring(1) : trimmedQuery;
  
  console.log('[userUtils] Создание User из SearchUserItem:', {
    id: user.id,
    name: user.name,
    username: user.us,
    query: query,
    isUsernameSearch,
    searchQuery,
  });

  const highlightText = (text: string, query: string) => {
    if (!query) return undefined;

    const lowerText = text.toLowerCase();
    const lowerQuery = query.toLowerCase();
    const index = lowerText.indexOf(lowerQuery);

    if (index === -1) return undefined;

    return {
      before: text.substring(0, index),
      match: text.substring(index, index + query.length),
      after: text.substring(index + query.length),
    };
  };

  let highlightedName = undefined;
  let highlightedUsername = undefined;

  if (isUsernameSearch) {
    const usernameHighlight = highlightText(user.us, searchQuery);
    if (usernameHighlight) {
      highlightedUsername = {
        before: `@${usernameHighlight.before}`,
        match: usernameHighlight.match,
        after: usernameHighlight.after,
      };
    }
  } else {
    highlightedName = highlightText(user.name, searchQuery);
  }

  console.log('[userUtils] Результат подсветки:', {
    highlightedName,
    highlightedUsername,
  });
  
  return {
    id: user.id.toString(),
    name: user.name,
    username: `@${user.us}`,
    description: user.status || '',
    imageUri: user.image,
    highlightedName,
    highlightedUsername,
    highlightedDescription: undefined,
  };
};