import {requestConstants} from '../_constants';

export function requests(state = {}, action) {
  switch (action.type) {
    case requestConstants.GET_ALL_REQUEST:
      return {...state, ...{
        adding: undefined,
        loading: true
      }};
    case requestConstants.GET_ALL_SUCCESS:
      return {...state, ...{
        items: action.requests.result,
        loading: false
      }};
    case requestConstants.GET_ALL_FAILURE:
      return {...state, ...{
        error: action.error,
        loading: false
      }};
    case requestConstants.ADD_REQUEST:
      return {...state, ...{
        adding: true
      }};
    case requestConstants.ADD_SUCCESS:
      return {...state, ...{
        adding: false
      }};
    case requestConstants.ADD_FAILURE:
      return {...state, ...{
        error: action.error
      }};
    case requestConstants.EDIT_REQUEST:
      return {...state, ...{
        adding: true
      }};
    case requestConstants.EDIT_SUCCESS:
      return {...state, ...{
        adding: false
      }};
    case requestConstants.EDIT_FAILURE:
      return {...state, ...{
        error: action.error
      }};
    case requestConstants.ACCEPT_REQUEST:
      return {...state, ...{
        loading: true
      }};
    case requestConstants.ACCEPT_SUCCESS:
      return {...state, ...{
        adding: false,
        loading: false
      }};
    case requestConstants.ACCEPT_FAILURE:
      return {...state, ...{
        error: action.error,
        loading: false
      }};
    case requestConstants.REJECT_REQUEST:
      return {...state, ...{
        loading: true
      }};
    case requestConstants.REJECT_SUCCESS:
      return {...state, ...{
        adding: false,
        loading: false
      }};
    case requestConstants.REJECT_FAILURE:
      return {...state, ...{
        error: action.error,
        loading: false
      }};
    case requestConstants.HISTORY_REQUEST:
      return {...state, ...{
        loading: true
      }};
    case requestConstants.HISTORY_SUCCESS:
      return {...state, ...{
        history: {
          [action.req.key.productKey]: action.history.result
        },
        loading: false
      }};
    case requestConstants.HISTORY_FAILURE:
      return {...state, ...{
        error: action.error,
        loading: false
      }};
    default:
      return state
  }
}