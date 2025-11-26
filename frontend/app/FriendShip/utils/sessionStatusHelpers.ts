export const getSessionStatus = (
  startTime: string, 
  duration: number
): 'recruitment' | 'in_progress' | 'completed' => {
  const now = new Date();

  let start: Date;
  
  if (startTime.includes('.')) {
    const parts = startTime.split(' ');
    const datePart = parts[0];
    const timePart = parts[1] || '00:00';
    
    const [day, month, year] = datePart.split('.');
    const [hour, minute] = timePart.split(':');
    
    start = new Date(
      parseInt(year), 
      parseInt(month) - 1,
      parseInt(day), 
      parseInt(hour) || 0, 
      parseInt(minute) || 0
    );
  } else {
    start = new Date(startTime);
  }

  const end = new Date(start.getTime() + duration * 60 * 1000);
  
  if (now < start) {
    return 'recruitment';
  } else if (now >= start && now < end) {
    return 'in_progress';
  } else {
    return 'completed';
  }
};

export const filterActiveSessions = (sessions: any[]): any[] => {
  return sessions.filter(session => {
    const status = getSessionStatus(
      session.session.start_time, 
      session.session.duration
    );
    return status === 'recruitment' || status === 'in_progress';
  });
};