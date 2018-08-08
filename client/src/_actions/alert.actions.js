import {alertConstants} from '../_constants';

const TIMEOUT = 5000;
export const alertActions = {
  success,
  error,
  clear
};

function success(message) {
  return dispatch => {
    setTimeout(() => {dispatch(clear())}, TIMEOUT);
    return dispatch({type: alertConstants.SUCCESS, message});
  };
}

function error(message) {
  return dispatch => {
    setTimeout(() => {dispatch(clear())}, TIMEOUT);
    return dispatch({type: alertConstants.ERROR, message});
  };
}

function clear() {
  return {type: alertConstants.CLEAR};
}