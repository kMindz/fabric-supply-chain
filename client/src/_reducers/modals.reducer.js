export function modals(state = {}, action) {
  const {modalId, type, ...rest} = action;
  if (type.startsWith('MODAL')) {
    return {
      ...state,
      ...{
        [modalId]: {
          ...rest,
          ...{
            show: type === 'MODAL_SHOW'
          }
        }
      }
    };
  }
  return state;
}