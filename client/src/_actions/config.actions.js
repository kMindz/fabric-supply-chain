import {configConstants} from '../_constants';
import {configService} from '../_services';

export const configActions = {
  get
};

function get() {
  return dispatch => {
    configService.load()
      .then(
         config => {
          dispatch(success(config));
        },
        error => {
          dispatch(failure(error.toString()));
        }
      );
  };

  function success(config) {
    return {type: configConstants.SUCCESS, config}
  }
  function failure(error) {
    return {type: configConstants.FAILURE, error}
  }
}