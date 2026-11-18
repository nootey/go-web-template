export class ApiError extends Error {
    status: number;
    data: unknown;
    isNetworkError: boolean;

    constructor(message: string, status: number, data: unknown, isNetworkError = false) {
        super(message);
        this.name = "ApiError";
        this.status = status;
        this.data = data;
        this.isNetworkError = isNetworkError;
    }
}

export type RequestOptions = {
    params?: Record<string, unknown>;
    headers?: Record<string, string>;
    responseType?: "blob";
};

// One page of a list. next_cursor is null on the last page.
export interface Page<T> {
    data: T[];
    next_cursor: string | null;
}
