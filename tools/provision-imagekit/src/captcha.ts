/**
 * CapSolver Turnstile solver via REST API.
 * Docs: https://docs.capsolver.com/en/guide/captcha/cloudflare_turnstile/
 */

const API_BASE = "https://api.capsolver.com";

interface CreateTaskResponse {
  errorId: number;
  errorCode?: string;
  errorDescription?: string;
  taskId?: string;
}

interface GetTaskResultResponse {
  errorId: number;
  errorCode?: string;
  errorDescription?: string;
  status: "idle" | "processing" | "ready" | "failed";
  solution?: { token: string };
}

export interface SolveResult {
  success: boolean;
  token?: string;
  error?: string;
}

/**
 * Solve a Cloudflare Turnstile challenge via CapSolver.
 * Returns the cf-turnstile-response token to inject into the page.
 */
export async function solveTurnstile(
  apiKey: string,
  websiteURL: string,
  websiteKey: string,
): Promise<SolveResult> {
  // 1. Create task
  const createRes = await fetch(`${API_BASE}/createTask`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      clientKey: apiKey,
      task: {
        type: "AntiTurnstileTaskProxyLess",
        websiteURL,
        websiteKey,
      },
    }),
  });

  const createBody = (await createRes.json()) as CreateTaskResponse;
  if (createBody.errorId !== 0 || !createBody.taskId) {
    return {
      success: false,
      error: `createTask failed: ${createBody.errorCode} — ${createBody.errorDescription}`,
    };
  }

  const taskId = createBody.taskId;

  // 2. Poll for result (max ~120s)
  const maxAttempts = 40;
  const pollInterval = 3000;

  for (let i = 0; i < maxAttempts; i++) {
    await new Promise((r) => setTimeout(r, pollInterval));

    const resultRes = await fetch(`${API_BASE}/getTaskResult`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ clientKey: apiKey, taskId }),
    });

    const resultBody = (await resultRes.json()) as GetTaskResultResponse;

    if (resultBody.errorId !== 0) {
      return {
        success: false,
        error: `getTaskResult failed: ${resultBody.errorCode} — ${resultBody.errorDescription}`,
      };
    }

    if (resultBody.status === "ready" && resultBody.solution?.token) {
      return { success: true, token: resultBody.solution.token };
    }

    if (resultBody.status === "failed") {
      return { success: false, error: "Task failed on solver side" };
    }
  }

  return { success: false, error: "Timeout waiting for captcha solution" };
}
