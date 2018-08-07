import {authService} from '../_services/auth.service';
import {getToken} from './user-store';

function _parseMessage(input = '') {
  const [,detailedMsg] = input.replace(')', '').split('message: ');
  return detailedMsg;
}

export function handleResponse(response) {
  return response.json().then(data => {
    if (!response.ok) {
      if (response.status === 401) {
        return authService.obtainToken()
          .catch(Promise.reject);
      }

      const error = (data && data.message && _parseMessage(data.message)) || response.statusText;
      return Promise.reject(new Error(error));
    }

    return data;
  });
}

function authHeader() {
  let token = getToken();

  if (token) {
    return {'Authorization': 'Bearer ' + token};
  } else {
    return {};
  }
}

export function getHeaders(headerObj = {}) {
  return {...headerObj, ...authHeader(), ...{'Content-Type': 'application/json'}}
}