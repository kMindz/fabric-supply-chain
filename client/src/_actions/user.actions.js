import {userConstants} from '../_constants';
import {userService} from '../_services';
import {alertActions} from './';
import {history} from '../_helpers';

export const userActions = {
  login,
  logout
};


function login(username, orgName) {
  return dispatch => {
    dispatch(request({username}));

    userService.login(username, orgName)
      .then(
        user => {
          dispatch(success(user));
          history.push('./');
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request(user) {
    return {type: userConstants.LOGIN_REQUEST, user}
  }

  function success(user) {
    return {type: userConstants.LOGIN_SUCCESS, user}
  }

  function failure(error) {
    return {type: userConstants.LOGIN_FAILURE, error}
  }
}

function logout() {
  return dispatch => {
    userService.logout();
    dispatch(success());
  };

  function success() {
    return {type: userConstants.LOGOUT};
  }
}