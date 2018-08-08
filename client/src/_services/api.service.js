import {getHeaders, sendRequest} from '../_helpers/request';
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
    method: 'GET'
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

  return sendRequest(url, requestOptions);
}

export function invoke(channel, chaincode, functionName, args) {
  const {org} = configService.get();
  const requestOptions = {
    method: 'POST',
    body: JSON.stringify({
      peers: [`${org}/peer0`],
      fcn: functionName,
      args
    })
  };

  return sendRequest(`/channels/${channel}/chaincodes/${chaincode}`, requestOptions);
}

export function login(user) {
  const requestOptions = {
    method: 'POST',
    body: JSON.stringify({username: user.name, orgName: user.org})
  };

  return sendRequest(`/users`, requestOptions);
}

export function config() {
  const requestOptions = {
    method: 'GET'
  };

  return sendRequest(`/config`, requestOptions);
}