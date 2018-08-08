export const modalActions = {
  register,
  show,
  hide
};

function register(modalId) {
  return {type: 'MODAL_HIDE', modalId};
}

function show(modalId, object) {
  return {type: 'MODAL_SHOW', modalId, object};
}

function hide(modalId) {
  return {type: 'MODAL_HIDE', modalId};
}