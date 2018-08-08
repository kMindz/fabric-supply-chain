import {modalConstants} from '../_constants';

export const modalActions = {
  register,
  show,
  hide,
  setData
};

function register(modalId) {
  return {type: modalConstants.HIDE, modalId};
}

function show(modalId, object) {
  return {type: modalConstants.SHOW, modalId, object};
}

function hide(modalId) {
  return {type: modalConstants.HIDE, modalId};
}

function setData(modalId, data) {
  return {type: modalConstants.SET_DATA, modalId, data};
}