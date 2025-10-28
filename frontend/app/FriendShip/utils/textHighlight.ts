export interface HighlightedText {
  before: string;
  match: string;
  after: string;
}

export const createHighlightedText = (text: string, query: string): HighlightedText | undefined => {
  if (!query.trim()) return undefined;

  const lowerText = text.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const index = lowerText.indexOf(lowerQuery);

  if (index === -1) return undefined;

  const beforeMatch = text.substring(0, index);
  const match = text.substring(index, index + query.length);
  const afterMatch = text.substring(index + query.length);

  return {
    before: beforeMatch,
    match: match,
    after: afterMatch
  };
};

export const highlightEventTitle = <T extends { title: string }>(
  item: T,
  query: string
): T & { highlightedTitle?: HighlightedText } => {
  const highlightedTitle = createHighlightedText(item.title, query);
  
  return {
    ...item,
    ...(highlightedTitle && { highlightedTitle })
  };
};

export const highlightGroupText = <T extends { name: string; description: string }>(
  item: T,
  query: string
): T & { highlightedName?: HighlightedText } => {
  const highlightedName = createHighlightedText(item.name, query);
  
  return {
    ...item,
    ...(highlightedName && { highlightedName }),
  };
};

export const highlightUserText = <T extends { name: string; username: string; description: string }>(
  item: T,
  query: string
): T & { 
  highlightedUsername?: HighlightedText;
} => {
  const highlightedUsername = createHighlightedText(item.username, query);
  
  return {
    ...item,
    ...(highlightedUsername && { highlightedUsername }),
  };
};