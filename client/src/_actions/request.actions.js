import {requestConstants} from '../_constants';
import {requestService} from '../_services';
import {alertActions} from './';

export const requestActions = {
  getAll,
  add,
  edit,
  accept,
  reject,
  history
};

function getAll() {
  return dispatch => {
    dispatch(request());
    requestService.getAll()
      .then(
         requests => {
          dispatch(success(requests));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: requestConstants.GET_ALL_REQUEST}
  }

  function success(requests) {
    return {type: requestConstants.GET_ALL_SUCCESS, requests}
  }

  function failure(error) {
    return {type: requestConstants.GET_ALL_FAILURE, error}
  }
}

function add(product, comment) {
  return dispatch => {
    dispatch(request());
    requestService.add(product, comment)
      .then(
        _ => {
          dispatch(success());
          dispatch(alertActions.success('Request was initiated'));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: requestConstants.ADD_REQUEST}
  }

  function success() {
    return {type: requestConstants.ADD_SUCCESS}
  }

  function failure(error) {
    return {type: requestConstants.ADD_FAILURE, error}
  }
}

function edit(req, comment) {
  return dispatch => {
    dispatch(request());
    requestService.edit(req, comment)
      .then(
        product => {
          dispatch(success());
          dispatch(alertActions.success('Request was updated'));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: requestConstants.EDIT_REQUEST}
  }

  function success() {
    return {type: requestConstants.EDIT_SUCCESS}
  }

  function failure(error) {
    return {type: requestConstants.EDIT_FAILURE, error}
  }
}


function accept(req) {
  return dispatch => {
    dispatch(request());
    requestService.accept(req)
      .then(
        _ => {
          dispatch(success());
          dispatch(alertActions.success('Request was accepted'));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: requestConstants.ACCEPT_REQUEST}
  }

  function success() {
    return {type: requestConstants.ACCEPT_SUCCESS}
  }

  function failure(error) {
    return {type: requestConstants.ACCEPT_FAILURE, error}
  }
}

function reject(req) {
  return dispatch => {
    dispatch(request());
    requestService.reject(req)
      .then(
        _ => {
          dispatch(success());
          dispatch(alertActions.success('Request was rejected'));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: requestConstants.REJECT_REQUEST}
  }

  function success() {
    return {type: requestConstants.REJECT_SUCCESS}
  }

  function failure(error) {
    return {type: requestConstants.REJECT_FAILURE, error}
  }
}

function history(req) {
  return dispatch => {
    dispatch(request());

    requestService.history(req)
      .then(
        history => {
          dispatch(success(req, history));
        },
        error => dispatch(failure(error.toString()))
      );
  };

  function request() {
    return {type: requestConstants.HISTORY_REQUEST};
  }

  function success(req, history) {
    return {type: requestConstants.HISTORY_SUCCESS, req, history};
  }

  function failure(error) {
    return {type: requestConstants.HISTORY_FAILURE, error};
  }
}