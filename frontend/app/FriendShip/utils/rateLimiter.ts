class RateLimiter {
  private actions: Map<string, number[]> = new Map();

  canPerformAction(key: string, maxActions: number, windowMs: number): boolean {
    const now = Date.now();
    const timestamps = this.actions.get(key) || [];

    const recentTimestamps = timestamps.filter(t => now - t < windowMs);
    
    if (recentTimestamps.length >= maxActions) {
      return false;
    }
    
    recentTimestamps.push(now);
    this.actions.set(key, recentTimestamps);
    return true;
  }

  reset(key: string): void {
    this.actions.delete(key);
  }

  resetAll(): void {
    this.actions.clear();
  }
}

export const rateLimiter = new RateLimiter();