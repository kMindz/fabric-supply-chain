const stages = ['request', 'success', 'failure'];
const actions = ['login', 'getall', 'add', 'edit', 'history'];

export const productConstants = {};
actions.forEach(action => {
  stages.forEach(stage => {
    const key = `${action.toUpperCase()}_${stage.toUpperCase()}`;
    productConstants[key] = 'PRODUCT_' + key;
  });
});

export const productStates = {
  1: 'Registered',
  2: 'Active',
  3: 'Decision-making',
  4: 'Inactive',
};
