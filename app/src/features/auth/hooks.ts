import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  changeEmail,
  changePassword,
  requestPasswordReset,
  resendVerificationEmail,
  resetPassword,
  signIn,
  signOut,
  signUp,
} from "@/features/auth/api";
import { accountKeys } from "@/features/account/query-keys";

export function useSignUp() {
  return useMutation({ mutationFn: signUp });
}

export function useSignIn() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: signIn,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: accountKeys.all });
    },
  });
}

export function useRequestPasswordReset() {
  return useMutation({ mutationFn: requestPasswordReset });
}

export function useResetPassword() {
  return useMutation({ mutationFn: resetPassword });
}

export function useSignOut() {
  return useMutation({ mutationFn: signOut });
}

export function useChangeEmail() {
  return useMutation({ mutationFn: changeEmail });
}

export function useChangePassword() {
  return useMutation({ mutationFn: changePassword });
}

export function useResendVerificationEmail() {
  return useMutation({ mutationFn: resendVerificationEmail });
}
