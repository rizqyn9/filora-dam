import fs from "node:fs/promises";
import path from "node:path";
import http from "node:http";
import { URL } from "node:url";
import { google } from "googleapis";

const ROOT = process.cwd();

const CREDENTIALS_PATH = path.join(ROOT, "credentials.json");

const TOKEN_PATH = path.join(ROOT, "token.json");

const SCOPES = ["https://www.googleapis.com/auth/gmail.readonly"];

let gmail = null;

/* -------------------------------------------------------
 * File helpers
 * ----------------------------------------------------- */

async function readJson(file) {
  try {
    const content = await fs.readFile(file, "utf8");
    return JSON.parse(content);
  } catch {
    return null;
  }
}

async function saveJson(file, data) {
  await fs.writeFile(file, JSON.stringify(data, null, 2), {
    encoding: "utf8",
    mode: 0o600,
  });
}

/* -------------------------------------------------------
 * OAuth client
 * ----------------------------------------------------- */

async function createOAuthClient() {
  const credentials = await readJson(CREDENTIALS_PATH);

  if (!credentials) {
    throw new Error(`Missing credentials.json:\n${CREDENTIALS_PATH}`);
  }

  const config = credentials.installed ?? credentials.web;

  if (!config) {
    throw new Error("credentials.json does not contain installed or web OAuth configuration.");
  }

  const { client_id, client_secret, redirect_uris = [] } = config;

  if (!client_id || !client_secret) {
    throw new Error("credentials.json is missing client_id/client_secret.");
  }

  /*
   * We use localhost because this is a local desktop OAuth client.
   * Google Desktop OAuth credentials normally include localhost
   * redirect URIs.
   */
  const redirectUri =
    redirect_uris.find((uri) => uri.startsWith("http://localhost")) ??
    "http://localhost:3000/oauth2callback";

  return new google.auth.OAuth2(client_id, client_secret, redirectUri);
}

/* -------------------------------------------------------
 * OAuth browser flow
 * ----------------------------------------------------- */

async function waitForAuthorizationCode(authUrl) {
  const redirectPort = 3000;

  const server = http.createServer(async (req, res) => {
    try {
      const requestUrl = new URL(req.url, `http://localhost:${redirectPort}`);

      if (requestUrl.pathname !== "/oauth2callback") {
        res.writeHead(404);
        res.end("Not found");
        return;
      }

      const error = requestUrl.searchParams.get("error");

      if (error) {
        res.writeHead(400);
        res.end(`OAuth error: ${error}`);

        server.close();

        throw new Error(`Google OAuth error: ${error}`);
      }

      const code = requestUrl.searchParams.get("code");

      if (!code) {
        res.writeHead(400);
        res.end("Missing authorization code.");
        return;
      }

      res.writeHead(200, {
        "Content-Type": "text/html; charset=utf-8",
      });

      res.end(`
            <!doctype html>
            <html>
              <body>
                <h2>Gmail authorization successful</h2>
                <p>You can close this browser tab.</p>
              </body>
            </html>
          `);

      server.close();

      return code;
    } catch (error) {
      console.error(error);
    }
  });

  await new Promise((resolve, reject) => {
    server.once("error", reject);

    server.listen(redirectPort, "127.0.0.1", resolve);
  });

  console.log("\nOpen this URL in your browser:\n");

  console.log(authUrl);
  console.log();

  /*
   * The server callback returns the code.
   *
   * Because Node's HTTP callback cannot directly resolve the
   * outer promise, we create a second promise below.
   */
}

/* -------------------------------------------------------
 * More reliable OAuth callback
 * ----------------------------------------------------- */

async function getAuthorizationCode(authUrl) {
  const port = 3000;

  return new Promise(async (resolve, reject) => {
    const server = http.createServer((req, res) => {
      try {
        const url = new URL(req.url, `http://localhost:${port}`);

        if (url.pathname !== "/oauth2callback") {
          res.writeHead(404);
          res.end("Not found");
          return;
        }

        const error = url.searchParams.get("error");

        if (error) {
          res.writeHead(400);
          res.end(`OAuth error: ${error}`);

          server.close();
          reject(new Error(`OAuth error: ${error}`));

          return;
        }

        const code = url.searchParams.get("code");

        if (!code) {
          res.writeHead(400);
          res.end("Authorization code missing.");

          server.close();

          reject(new Error("Authorization code missing."));

          return;
        }

        res.writeHead(200, {
          "Content-Type": "text/html; charset=utf-8",
        });

        res.end(`
                <!doctype html>
                <html>
                  <head>
                    <title>Gmail Authorized</title>
                  </head>
                  <body>
                    <h2>Gmail authorization successful.</h2>
                    <p>You can close this tab.</p>
                  </body>
                </html>
              `);

        server.close();

        resolve(code);
      } catch (error) {
        server.close();
        reject(error);
      }
    });

    server.on("error", reject);

    server.listen(port, "127.0.0.1", () => {
      console.log("\nOpen this URL in your browser:\n");

      console.log(authUrl);

      console.log("\nWaiting for Google OAuth callback...");
    });
  });
}

/* -------------------------------------------------------
 * Authorize
 * ----------------------------------------------------- */

export async function authorize() {
  if (gmail) {
    return gmail;
  }

  const auth = await createOAuthClient();

  const savedToken = await readJson(TOKEN_PATH);

  if (savedToken) {
    console.log("Loading saved Gmail token...");

    auth.setCredentials(savedToken);

    /*
     * Save refreshed credentials whenever Google
     * refreshes the access token.
     */
    auth.on("tokens", async (tokens) => {
      const current = (await readJson(TOKEN_PATH)) ?? {};

      const updated = {
        ...current,
        ...tokens,
      };

      await saveJson(TOKEN_PATH, updated);

      console.log("Gmail token refreshed.");
    });

    gmail = google.gmail({
      version: "v1",
      auth,
    });

    /*
     * Actually verify the credentials.
     */
    await gmail.users.getProfile({
      userId: "me",
    });

    return gmail;
  }

  console.log("No token.json found.");

  console.log("Starting Gmail OAuth authorization...");

  const authUrl = auth.generateAuthUrl({
    access_type: "offline",
    prompt: "consent",
    scope: SCOPES,
  });

  const code = await getAuthorizationCode(authUrl);

  const { tokens } = await auth.getToken(code);

  if (!tokens.refresh_token) {
    throw new Error(
      "Google did not return a refresh_token. Delete token.json if it exists and authorize again.",
    );
  }

  auth.setCredentials(tokens);

  await saveJson(TOKEN_PATH, tokens);

  console.log(`Saved Gmail token to:\n${TOKEN_PATH}`);

  gmail = google.gmail({
    version: "v1",
    auth,
  });

  /*
   * Verify immediately.
   */
  await gmail.users.getProfile({
    userId: "me",
  });

  return gmail;
}

/* -------------------------------------------------------
 * Profile
 * ----------------------------------------------------- */

export async function getProfile() {
  const client = await authorize();

  const response = await client.users.getProfile({
    userId: "me",
  });

  return response.data;
}

/* -------------------------------------------------------
 * List message IDs
 * ----------------------------------------------------- */

export async function listMessages({
  query = "",
  maxResults = 20,
  pageToken,
  includeSpamTrash = false,
} = {}) {
  const client = await authorize();

  const response = await client.users.messages.list({
    userId: "me",
    q: query || undefined,
    maxResults,
    pageToken,
    includeSpamTrash,
  });

  return {
    messages: response.data.messages ?? [],

    nextPageToken: response.data.nextPageToken ?? null,

    resultSizeEstimate: response.data.resultSizeEstimate ?? 0,
  };
}

/* -------------------------------------------------------
 * Get complete message
 * ----------------------------------------------------- */

export async function getMessage(messageId, { format = "full" } = {}) {
  if (!messageId) {
    throw new Error("messageId is required.");
  }

  const client = await authorize();

  const response = await client.users.messages.get({
    userId: "me",
    id: messageId,
    format,
  });

  return response.data;
}

/* -------------------------------------------------------
 * Headers
 * ----------------------------------------------------- */

export function getHeaders(message) {
  const headers = message?.payload?.headers ?? [];

  const result = {};

  for (const header of headers) {
    if (!header.name) {
      continue;
    }

    result[header.name.toLowerCase()] = header.value ?? "";
  }

  return result;
}

export function getHeader(message, name) {
  const headers = message?.payload?.headers ?? [];

  const target = name.toLowerCase();

  const header = headers.find((header) => header.name?.toLowerCase() === target);

  return header?.value ?? null;
}

/* -------------------------------------------------------
 * Body decoding
 * ----------------------------------------------------- */

function decodeBase64Url(data) {
  if (!data) {
    return "";
  }

  const normalized = data.replace(/-/g, "+").replace(/_/g, "/");

  const padding = normalized.length % 4;

  const padded = padding === 0 ? normalized : normalized + "=".repeat(4 - padding);

  return Buffer.from(padded, "base64").toString("utf8");
}

/* -------------------------------------------------------
 * Extract body recursively
 * ----------------------------------------------------- */

function collectParts(part, result) {
  if (!part) {
    return;
  }

  const mimeType = part.mimeType ?? "";

  if (mimeType === "text/plain" && part.body?.data) {
    result.text.push(decodeBase64Url(part.body.data));
  }

  if (mimeType === "text/html" && part.body?.data) {
    result.html.push(decodeBase64Url(part.body.data));
  }

  if (Array.isArray(part.parts)) {
    for (const child of part.parts) {
      collectParts(child, result);
    }
  }
}

/* -------------------------------------------------------
 * Get body
 * ----------------------------------------------------- */

export function getMessageBody(message) {
  const result = {
    text: [],
    html: [],
  };

  collectParts(message?.payload, result);

  return {
    text: result.text.join("\n"),
    html: result.html.join("\n"),
  };
}

/* -------------------------------------------------------
 * Normalize message
 * ----------------------------------------------------- */

export function normalizeMessage(message) {
  const headers = getHeaders(message);

  const body = getMessageBody(message);

  return {
    id: message.id,
    threadId: message.threadId,

    from: headers.from ?? null,

    to: headers.to ?? null,

    cc: headers.cc ?? null,

    subject: headers.subject ?? null,

    date: headers.date ?? null,

    messageId: headers["message-id"] ?? null,

    snippet: message.snippet ?? null,

    internalDate: message.internalDate ? Number(message.internalDate) : null,

    labels: message.labelIds ?? [],

    text: body.text,

    html: body.html,

    sizeEstimate: message.sizeEstimate ?? null,
  };
}

/* -------------------------------------------------------
 * Search
 * ----------------------------------------------------- */

export async function searchMessages({ query, maxResults = 20, includeSpamTrash = false } = {}) {
  if (!query) {
    throw new Error("query is required.");
  }

  const result = await listMessages({
    query,
    maxResults,
    includeSpamTrash,
  });

  const messages = [];

  for (const item of result.messages) {
    const message = await getMessage(item.id, {
      format: "full",
    });

    messages.push(normalizeMessage(message));
  }

  return {
    messages,
    nextPageToken: result.nextPageToken,
  };
}

/* -------------------------------------------------------
 * Wait for message
 * ----------------------------------------------------- */

export async function waitForMessage({ query, timeoutMs = 120000, intervalMs = 2000 } = {}) {
  if (!query) {
    throw new Error("query is required.");
  }

  const started = Date.now();

  while (Date.now() - started < timeoutMs) {
    const result = await searchMessages({
      query,
      maxResults: 10,
    });

    if (result.messages.length) {
      return result.messages[0];
    }

    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  throw new Error(`Timeout waiting for Gmail message: ${query}`);
}
