export function getRecapTotemStage(activeIndex: number, totalCards: number) {
  if (totalCards <= 1) return 4;
  const ratio = activeIndex / (totalCards - 1);
  if (ratio < 0.2) return 0;
  if (ratio < 0.42) return 1;
  if (ratio < 0.66) return 2;
  if (ratio < 0.86) return 3;
  return 4;
}
