import dayjs from 'dayjs';

export const formatDate = (dateString: string): string => {

  if (dateString.includes('.') && dateString.includes(' ')) {
    return dateString.split(' ')[0];
  }

  if (dateString.includes('T') || dateString.includes('Z')) {
    return dayjs(dateString).format('DD.MM.YYYY');
  }

  return dateString;
};
