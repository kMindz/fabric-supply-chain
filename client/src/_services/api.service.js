import {getHeaders, handleResponse} from '../_helpers/request';
import {configService} from './config.service';

//TODO get from config
export const contracts = {
  reference: 'reference',
  relationship: 'relationship'
};
export const channels = {
  common: 'common',
  ab: 'a-b',
  ac: 'a-c',
  bc: 'b-c'
};

export function query(channel, chaincode, fcn, args) {
  const requestOptions = {
    method: 'GET',
    headers: getHeaders()
  };

  const {org} = configService.get();
  const params = {
    peer: `${org}/peer0`,
    fcn,
    args
  };

  //TODO set host
  const url = new URL(`${window.location.origin}/channels/${channel}/chaincodes/${chaincode}`);
  Object.keys(params).forEach(key => url.searchParams.append(key, params[key]));

  return fetch(url, requestOptions)
    .then(handleResponse);
}

export function invoke(channel, chaincode, functionName, args) {
  const {org} = configService.get();
  const requestOptions = {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({
      peers: [`${org}/peer0`],
      fcn: functionName,
      args
    })
  };

  return fetch(`/channels/${channel}/chaincodes/${chaincode}`, requestOptions)
    .then(handleResponse);
}

export function login(user) {
  const requestOptions = {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({username: user.name, orgName: user.org})
  };

  return fetch(`/users`, requestOptions)
    .then(handleResponse);
}

export function config() {
  const requestOptions = {
    method: 'GET',
    headers: {'Content-Type': 'application/json'}
  };

  return fetch(`/config`, requestOptions)
    .then(handleResponse);
}