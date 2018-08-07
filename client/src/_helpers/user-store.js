const userKey = 'user';
const tokenKey = 'token';

export function set(user) {
  localStorage.setItem(userKey, JSON.stringify(user));
  localStorage.setItem(tokenKey, user.token);
}

export function get() {
  const data = localStorage.getItem(userKey);
  if (data) {
    return JSON.parse(data);
  }
  return null;
}

export function clear() {
  localStorage.removeItem(userKey);
  localStorage.removeItem(tokenKey);
}

export function getToken() {
  return localStorage.getItem(tokenKey);
}