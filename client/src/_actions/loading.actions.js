export const loadingActions = {
  start,
  end
};

function start() {
  return {type: 'LOADING_START'};
}

function end() {
  return {type: 'LOADING_END'};
}