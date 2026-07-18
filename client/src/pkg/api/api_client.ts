import { useAuthStore } from "../stores/auth_store.ts";
import { ApiError } from "./api_models.ts";
import type { RequestOptions } from "./api_models.ts";

function appendParam(searchParams: URLSearchParams, key: string, value: unknown): void {
    if (value === null || value === undefined) return;
    if (Array.isArray(value)) {
        value.forEach((item, i) => appendParam(searchParams, `${key}[${i}]`, item));
    } else if (typeof value === "object") {
        Object.entries(value as Record<string, unknown>).forEach(([k, v]) =>
            appendParam(searchParams, `${key}[${k}]`, v),
        );
    } else {
        searchParams.append(key, String(value));
    }
}

function buildUrl(path: string, params?: Record<string, unknown>): string {
    const url = new URL(`/api/${path}`, window.location.origin);
    if (params) {
        Object.entries(params).forEach(([k, v]) => appendParam(url.searchParams, k, v));
    }
    return url.pathname + url.search;
}

async function doFetch(method: string, path: string, body: unknown, options: RequestOptions): Promise<Response> {
    const url = buildUrl(path, options.params);
    const headers: Record<string, string> = { ...options.headers };
    let fetchBody: BodyInit | undefined;

    if (body instanceof FormData) {
        fetchBody = body;
    } else if (body instanceof Blob) {
        fetchBody = body;
        if (body.type) headers["Content-Type"] = body.type;
    } else if (body !== null && body !== undefined) {
        headers["Content-Type"] = "application/json";
        fetchBody = JSON.stringify(body);
    }

    return fetch(url, {
        method,
        headers,
        body: fetchBody,
        credentials: "include",
    });
}

async function handleResponse<T>(response: Response, options: RequestOptions): Promise<{ data: T }> {
    if (options.responseType === "blob") {
        return { data: (await response.blob()) as T };
    }
    const text = await response.text();
    return { data: (text ? JSON.parse(text) : null) as T };
}

// A 401 from these is an expected outcome the caller handles itself (bad
// credentials, an already-dead session), not a signal to tear down auth state.
const authEndpointPattern = /\/auth\/(login|register|logout|me)/;

async function request<T>(method: string, path: string, body: unknown, options: RequestOptions): Promise<{ data: T }> {
    let response: Response;

    try {
        response = await doFetch(method, path, body, options);
    } catch {
        throw new ApiError("Network Error", 0, null, true);
    }

    if (!response.ok) {
        let data: unknown = null;
        try {
            data = await response.json();
        } catch {
            // ignore parse errors
        }

        if (response.status === 401 && !authEndpointPattern.test(`/${path}`)) {
            const auth = useAuthStore();
            if (auth.isAuthenticated) {
                await auth.logoutUser();
            }
        }

        const msg = (data as { message?: string } | null)?.message ?? response.statusText;
        throw new ApiError(msg, response.status, data);
    }

    return handleResponse<T>(response, options);
}

const apiClient = {
    get: <T = any>(path: string, options?: RequestOptions) => request<T>("GET", path, undefined, options ?? {}),
    post: <T = any>(path: string, body?: unknown, options?: RequestOptions) =>
        request<T>("POST", path, body, options ?? {}),
    put: <T = any>(path: string, body?: unknown, options?: RequestOptions) =>
        request<T>("PUT", path, body, options ?? {}),
    patch: <T = any>(path: string, body?: unknown, options?: RequestOptions) =>
        request<T>("PATCH", path, body, options ?? {}),
    delete: <T = any>(path: string, options?: RequestOptions) => request<T>("DELETE", path, undefined, options ?? {}),
};

export default apiClient;
