import {combineReducers} from 'redux';

import {authentication} from './authentication.reducer';
import {alert} from './alert.reducer';
import {config} from './config.reducer';
import {loading} from './loading.reducer';
import {products} from './products.reducer';
import {requests} from './requests.reducer';
import {modals} from './modals.reducer';

const rootReducer = combineReducers({
  authentication,
  alert,
  loading,
  products,
  config,
  requests,
  modals
});

export default rootReducer;