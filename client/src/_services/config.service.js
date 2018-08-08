import {config} from './api.service';

export const configService = {
  load,
  get
};

let localConfig = {};
let isFetching = false;
let reqPromise;

function get() {
  return localConfig;
}

function load() {
  if (isFetching) {
    return reqPromise;
  }
  isFetching = true;

  reqPromise = config()
    .then(res => {
      localConfig = res;

      //TODO extend config, load channels, chaincodes
      return localConfig;
    });

  return reqPromise;
}
