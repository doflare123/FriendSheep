import { SearchUserItem } from '@/api/types/user';
import { User } from '@/components/profile/UserCard';

export const createUserWithHighlightedText = (
  user: SearchUserItem,
  query: string
): User => {
  const highlightText = (text: string, query: string) => {
    if (!query.trim()) return undefined;

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
  
  const usernameWithAt = `@${user.us}`;
  
  console.log('[userUtils] Создание User из SearchUserItem:', {
    id: user.id,
    name: user.name,
    username: user.us,
    image: user.image,
    imageEmpty: !user.image || user.image.trim() === '',
  });

  const highlightedName = highlightText(user.name, query);

  const highlightedUsername = highlightText(user.us, query);
  
  return {
    id: user.id.toString(),
    name: user.name,
    username: usernameWithAt,
    description: user.status || '',
    imageUri: user.image,
    highlightedUsername: highlightedUsername ? {
      before: '@' + highlightedUsername.before,
      match: highlightedUsername.match,
      after: highlightedUsername.after,
    } : undefined,
    highlightedName: highlightedName,
    highlightedDescription: undefined,
  };
};