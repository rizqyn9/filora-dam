const adjectives = [
  "bright", "swift", "calm", "bold", "warm", "keen", "fair", "wise",
  "glad", "pure", "soft", "cool", "wild", "free", "true", "blue",
  "red", "dark", "light", "deep", "high", "low", "new", "old",
  "quick", "slow", "sharp", "flat", "round", "long", "short", "big",
];

const nouns = [
  "fox", "owl", "wolf", "bear", "deer", "hawk", "dove", "crow",
  "lake", "pine", "oak", "elm", "ivy", "fern", "moss", "reed",
  "stone", "cliff", "brook", "dale", "glen", "vale", "ridge", "peak",
  "cloud", "star", "moon", "rain", "wind", "snow", "dusk", "dawn",
];

function pick<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

export function generateEmail(domain: string): string {
  const adj = pick(adjectives);
  const noun = pick(nouns);
  const num = Math.floor(Math.random() * 900) + 100;
  return `${adj}${noun}${num}@${domain}`;
}
