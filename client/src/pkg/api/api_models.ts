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
