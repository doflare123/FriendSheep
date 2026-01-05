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