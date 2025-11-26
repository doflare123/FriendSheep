export function extractCityFromAddress(address: string): string {
  if (!address || !address.trim()) return '';
  
  const cleaned = address.trim();

  const cityPrefixMatch = cleaned.match(/^(?:г\.\s*|город\s+)([^,]+)/i);
  if (cityPrefixMatch) {
    return cityPrefixMatch[1].trim();
  }

  const firstPart = cleaned.split(',')[0].trim();

  const notCityPrefixes = /^(ул\.|улица|пр\.|проспект|пер\.|переулок|д\.|дом|кв\.|квартира)/i;
  if (!notCityPrefixes.test(firstPart)) {
    return firstPart;
  }
  
  return '';
}