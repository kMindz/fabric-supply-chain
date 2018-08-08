import * as apiService from './api.service';
import {configService} from './config.service';

export const productService = {
  getAll,
  add,
  edit,
  history
};

function getAll() {
  return apiService.query(
    apiService.channels.common,
    apiService.contracts.reference,
    'queryProducts',
    `["${encodeURI(JSON.stringify({
      'selector': {
        'docType': 'product'/*,
        'owner': {
          '$in': ['a', 'b', 'c']
        }*/
      }
    }))}"]`);
}

function add(product) {
  const {org} = configService.get();
  return apiService.invoke(
    apiService.channels.common,
    apiService.contracts.reference,
    'initProduct',
    [product.name, product.desc, '1' /*initial state*/, org, Date.now() + '']
  );
}

function edit(product) {
  const {org} = configService.get();
  return apiService.invoke(
    apiService.channels.common,
    apiService.contracts.reference,
    'updateProduct',
    [product.name, product.desc, product.state + '', org, Date.now() + '']
  );
}

function history(product) {
  return apiService.query(
    apiService.channels.common,
    apiService.contracts.reference,
    'getHistoryForProduct',
    `["${product.key.name}"]`
  );
}
