export function loading(state = false, action) {
  const { type } = action;
  const startMatches = /(.*)_(REQUEST)/.exec(type);
  const endMatches = /(.*)_(SUCCESS|FAILURE)/.exec(type);

  if (!startMatches && !endMatches) return state;

  return !!startMatches || !endMatches;
}
