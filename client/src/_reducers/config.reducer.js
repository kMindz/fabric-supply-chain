import {configConstants} from '../_constants';

export function config(state = null, action) {
  switch (action.type) {
    case configConstants.SUCCESS:
      return action.config;
    default:
      return state
  }
}