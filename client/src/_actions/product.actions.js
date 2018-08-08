import {productConstants} from '../_constants';
import {productService} from '../_services';
import {alertActions} from './';

export const productActions = {
  add,
  edit,
  getAll,
  history
};

function add(product) {
  return dispatch => {
    dispatch(request());
    productService.add(product)
      .then(
        product => {
          dispatch(success());
          dispatch(alertActions.success('Product was added'));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: productConstants.ADD_REQUEST}
  }

  function success() {
    return {type: productConstants.ADD_SUCCESS}
  }

  function failure(error) {
    return {type: productConstants.ADD_FAILURE, error}
  }
}

function edit(product) {
  return dispatch => {
    dispatch(request());
    productService.edit(product)
      .then(
        product => {
          dispatch(success());
          dispatch(alertActions.success('Product was updated'));
        },
        error => {
          dispatch(failure(error.toString()));
          dispatch(alertActions.error(error.toString()));
        }
      );
  };

  function request() {
    return {type: productConstants.EDIT_REQUEST}
  }

  function success() {
    return {type: productConstants.EDIT_SUCCESS}
  }

  function failure(error) {
    return {type: productConstants.EDIT_FAILURE, error}
  }
}

function getAll() {
  return dispatch => {
    dispatch(request());

    productService.getAll()
      .then(
        products => {
          dispatch(success(products));
        },
        error => dispatch(failure(error.toString()))
      );
  };

  function request() {
    return {type: productConstants.GETALL_REQUEST}
  }

  function success(products) {
    return {type: productConstants.GETALL_SUCCESS, products}
  }

  function failure(error) {
    return {type: productConstants.GETALL_FAILURE, error}
  }
}

function history(product) {
  return dispatch => {
    dispatch(request());

    productService.history(product)
      .then(
        history => {
          dispatch(success(product, history));
        },
        error => dispatch(failure(error.toString()))
      );
  };

  function request() {
    return {type: productConstants.HISTORY_REQUEST};
  }

  function success(product, history) {
    return {type: productConstants.HISTORY_SUCCESS, product, history};
  }

  function failure(error) {
    return {type: productConstants.HISTORY_FAILURE, error};
  }
}