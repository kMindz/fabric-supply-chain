import {productConstants} from '../_constants';

export function products(state = {items: []}, action) {
  switch (action.type) {
    case productConstants.GETALL_REQUEST:
      return {...state, ...{
        loading: true,
        adding: undefined
      }};
    case productConstants.GETALL_SUCCESS:
      return {...state, ...{
        items: action.products.result,
        loading: false
      }};
    case productConstants.GETALL_FAILURE:
      return {...state, ...{
        error: action.error,
        loading: false
      }};
    case productConstants.ADD_REQUEST:
      return {...state, ...{
        adding: true
      }};
    case productConstants.ADD_SUCCESS:
      return {...state, ...{
        adding: false
      }};
    case productConstants.ADD_FAILURE:
      return {...state, ...{
        error: action.error
      }};
    case productConstants.EDIT_REQUEST:
      return {...state, ...{
        adding: true
      }};
    case productConstants.EDIT_SUCCESS:
      return {...state, ...{
        adding: false
      }};
    case productConstants.EDIT_FAILURE:
      return {...state, ...{
        error: action.error
      }};
    case productConstants.HISTORY_REQUEST:
      return {...state, ...{
        loading: true
      }};
    case productConstants.HISTORY_SUCCESS:
      return {...state, ...{
        history: {
          [action.product.key.name]: action.history.result
        },
        loading: false
      }};
    case productConstants.HISTORY_FAILURE:
      return {...state, ...{
        error: action.error,
        loading: false
      }};

    default:
      return state;
  }
}