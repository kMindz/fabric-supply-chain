import {set, get, clear} from '../_helpers';
import {login} from './api.service';

export const authService = {
  obtainToken
};

function obtainToken() {
  const user = get();
  if (!user) {
    clear();
    window.location.reload(true);
  }
  return login(user)
    .then(res => {
      if (res.token) {
        user.token = res.token;
        set(user);
      }

      return user;
    });
}
