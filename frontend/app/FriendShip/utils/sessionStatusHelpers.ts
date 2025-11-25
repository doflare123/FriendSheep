export const getSessionStatus = (
  startTime: string, 
  duration: number
): 'recruitment' | 'in_progress' | 'completed' => {
  const now = new Date();
  const start = new Date(startTime);

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