// HTML5のtype="email"相当の実用的なメールアドレス判定。
export const EMAIL_REGEX =
  /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

// 大文字・小文字・数字をそれぞれ1文字以上含み、8文字以上であることを要求する。
// (Supabase/bcryptの実用上の上限として72文字までに制限する。)
export const PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,72}$/;

export function validateEmail(email: string): string | undefined {
  if (!email) {
    return "メールアドレスを入力してください";
  }
  if (!EMAIL_REGEX.test(email)) {
    return "メールアドレスの形式が正しくありません";
  }
  return undefined;
}

export function validatePassword(password: string): string | undefined {
  if (!password) {
    return "パスワードを入力してください";
  }
  if (!PASSWORD_REGEX.test(password)) {
    return "パスワードは大文字・小文字・数字をそれぞれ1文字以上含む8文字以上で入力してください";
  }
  return undefined;
}

export function validateRequired(value: string, message: string): string | undefined {
  return value ? undefined : message;
}

export function validatePasswordConfirmation(
  password: string,
  confirmPassword: string,
): string | undefined {
  if (!confirmPassword) {
    return "確認用のパスワードを入力してください";
  }
  if (password !== confirmPassword) {
    return "パスワードが一致しません";
  }
  return undefined;
}
