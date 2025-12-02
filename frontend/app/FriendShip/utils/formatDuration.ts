export const formatDuration = (minutes: number | string): string => {
  const mins = typeof minutes === 'string' ? parseInt(minutes) : minutes;
  
  if (isNaN(mins)) return '0 мин';
  
  const hours = Math.floor(mins / 60);
  const remainingMins = mins % 60;
  
  if (hours === 0) {
    return `${mins} мин`;
  }
  
  if (remainingMins === 0) {
    return `${hours} ч`;
  }
  
  return `${hours} ч ${remainingMins} мин`;
};