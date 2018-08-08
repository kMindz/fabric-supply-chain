import * as apiService from './api.service';
import {configService} from './config.service';

export const requestService = {
  getAll,
  add,
  edit,
  accept,
  reject,
  history
};

function _selectChannelFromProduct(product) {
  return _selectChannel(configService.get().org, product.value.owner);
}
function _selectChannelFromRequest(request) {
  const {requestSender, requestReceiver} = request.key;
  return _selectChannel(requestSender, requestReceiver);
}
function _selectChannel(org1, org2) {
  const [first, second] = [org1, org2].sort();
  return apiService.channels[`${first}${second}`];
}
function _getChannels() {
  const userOrg = configService.get().org;
  return Object.entries(apiService.channels)
    .filter(([k, v]) => {
      return k.includes(userOrg) && v.includes('-');
    })
    .map(([k, v]) => v);
}

function getAll() {
  return Promise.all(_getChannels().map(ch => {
    return apiService.query(
      ch,
      apiService.contracts.relationship,
      'query',
      `[]`);
  }))
    .then(res => {
      return {result: res.reduce((acc, currentValue) => {
        return [...acc, ...currentValue.result];
      }, [])};
    });
}

function add(product, comment) {
  const {org} = configService.get();
  return apiService.invoke(
    _selectChannelFromProduct(product),
    apiService.contracts.relationship,
    'sendRequest',
    [product.key.name, org, product.value.owner, comment]
  );
}

function edit(request, comment) {
  const {org} = configService.get();
  return apiService.invoke(
    _selectChannelFromRequest(request),
    apiService.contracts.relationship,
    'editRequest',
    [request.key.productKey, org, request.key.requestReceiver, comment]
  );
}

function accept(request) {
  return apiService.invoke(
    _selectChannelFromRequest(request),
    apiService.contracts.relationship,
    'transferAccepted',
    [request.key.productKey, request.key.requestSender, request.key.requestReceiver]
  );
}

function reject(request) {
  return apiService.invoke(
    _selectChannelFromRequest(request),
    apiService.contracts.relationship,
    'transferRejected',
    [request.key.productKey, request.key.requestSender, request.key.requestReceiver]
  );
}

function history(request) {
  return apiService.query(
    _selectChannelFromRequest(request),
    apiService.contracts.relationship,
    'history',
    `["${request.key.productKey}"]`
  );
}
