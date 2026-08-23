import pino from "pino";

// XTERM 256 colors — visually distinct, readable on dark backgrounds
const COLORS = [
  "\x1b[38;5;39m", // blue
  "\x1b[38;5;208m", // orange
  "\x1b[38;5;41m", // green
  "\x1b[38;5;171m", // purple
  "\x1b[38;5;214m", // gold
  "\x1b[38;5;81m", // cyan
  "\x1b[38;5;204m", // pink
  "\x1b[38;5;118m", // lime
  "\x1b[38;5;147m", // lavender
  "\x1b[38;5;216m", // peach
];
const RESET = "\x1b[0m";

const colorMap = new Map<string, string>();
let colorIdx = 0;

function getColor(accountId: string): string {
  if (!colorMap.has(accountId)) {
    colorMap.set(accountId, COLORS[colorIdx % COLORS.length]);
    colorIdx++;
  }
  return colorMap.get(accountId)!;
}

export const rootLogger = pino({
  level: "debug",
  transport: {
    target: "pino-pretty",
    options: {
      colorize: true,
      ignore: "pid,hostname",
      translateTime: "HH:MM:ss",
    },
  },
});

export function createAccountLogger(accountId: string, email: string) {
  const color = getColor(accountId);
  const prefix = `${color}[${accountId}|${email}]${RESET}`;

  return {
    info(msg: string) {
      rootLogger.info(`${prefix} ${msg}`);
    },
    warn(msg: string) {
      rootLogger.warn(`${prefix} ${msg}`);
    },
    error(msg: string) {
      rootLogger.error(`${prefix} ${msg}`);
    },
    debug(msg: string) {
      rootLogger.debug(`${prefix} ${msg}`);
    },
    success(msg: string) {
      rootLogger.info(`${prefix} \x1b[32m✓${RESET} ${msg}`);
    },
    fail(msg: string) {
      rootLogger.error(`${prefix} \x1b[31m✗${RESET} ${msg}`);
    },
    state(from: string, to: string) {
      rootLogger.info(`${prefix} ${from} → \x1b[1m${to}${RESET}`);
    },
  };
}

export type AccountLogger = ReturnType<typeof createAccountLogger>;
