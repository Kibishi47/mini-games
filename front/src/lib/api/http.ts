const API_URL = import.meta.env.VITE_API_URL ?? "";

type HttpOptions = RequestInit & {
  json?: unknown;
};

export class HttpError extends Error {
  status: number;
  body: unknown;

  constructor(status: number, body: unknown, message = "HTTP Error") {
    super(message);
    this.status = status;
    this.body = body;
  }
}

export async function http<T>(path: string, options: HttpOptions = {}): Promise<T> {
  const { json, headers, ...rest } = options;

  const res = await fetch(`${API_URL}${path}`, {
    ...rest,
    headers: {
      ...(json ? { "Content-Type": "application/json" } : {}),
      ...(headers ?? {}),
    },
    body: json ? JSON.stringify(json) : rest.body,
    // IMPORTANT: indispensable pour envoyer/recevoir les cookies httpOnly (refresh token)
    credentials: "include",
  });

  // Essaie de parser le body
  const contentType = res.headers.get("content-type") ?? "";
  const isJson = contentType.includes("application/json");
  const body = isJson ? await res.json().catch(() => null) : await res.text().catch(() => null);

  if (!res.ok) {
    throw new HttpError(res.status, body, `HTTP ${res.status}`);
  }

  return body as T;
}