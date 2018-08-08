import {authService} from '../_services/auth.service';
import {getToken} from './user-store';

function _parseMessage(input = '') {
  const [,detailedMsg] = input.replace(')', '').split('message: ');
  return detailedMsg;
}

export function sendRequest(url, options) {
  options.headers = getHeaders(options.headers);
  return fetch(url, options)
    .then(handleResponse)
    .catch(e => {
      if (e.status === 401) {
        //try to resend a request
        return authService.obtainToken()
          .then(() => {
            options.headers = getHeaders(options.headers);
            return fetch(url, options);
          })
          .then(handleResponse);
      }
      return Promise.reject(e);
    });
}

function handleResponse(response) {
  return response.json().then(data => {
    if (!response.ok) {
      if (response.status === 401) {
        return Promise.reject(response);
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