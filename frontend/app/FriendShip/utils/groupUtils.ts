import { Group } from '@/components/groups/GroupCard';

export const createGroupWithHighlightedName = (
  group: Group,
  query: string
): Group => {
  if (!query.trim()) return group;

  const name = group.name;
  const lowerName = name.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const index = lowerName.indexOf(lowerQuery);

  if (index === -1) return group;

  const beforeMatch = name.substring(0, index);
  const match = name.substring(index, index + query.length);
  const afterMatch = name.substring(index + query.length);

  return {
    ...group,
    highlightedName: {
      before: beforeMatch,
      match: match,
      after: afterMatch,
    },
  };
};