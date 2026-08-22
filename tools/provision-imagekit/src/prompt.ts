import { createInterface } from "readline";

const rl = createInterface({ input: process.stdin, output: process.stdout });

export function ask(question: string): Promise<string> {
  return new Promise((resolve) => {
    rl.question(question, (answer) => resolve(answer.trim()));
  });
}

export async function pressEnter(message: string): Promise<void> {
  await ask(message + " ");
}

export function closePrompt(): void {
  rl.close();
}
