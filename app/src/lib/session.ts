export class SessionExpiredError extends Error {
  constructor() {
    super('SESSION_EXPIRED');
    this.name = 'SessionExpiredError';
  }
}

type Listener = () => void;

let listener: Listener | null = null;
let expired = false;

export function onSessionExpired(fn: Listener) {
  listener = fn;
  return () => {
    if (listener === fn) listener = null;
  };
}

export function notifySessionExpired() {
  if (expired) return;
  expired = true;
  listener?.();
}

export function isSessionExpired() {
  return expired;
}
