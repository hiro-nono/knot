// バックエンドが返すエラー(controller.writeError)はすべて {"error": string} の
// JSON形式なので、そこから抽出したメッセージとHTTPステータスを保持する。
export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isForbidden(): boolean {
    return this.status === 403;
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }
}

export function getApiErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return "予期しないエラーが発生しました";
}
