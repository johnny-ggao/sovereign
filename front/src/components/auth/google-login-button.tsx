"use client"

import { useRouter } from "next/navigation"
import { GoogleLogin } from "@react-oauth/google"
import { useGoogleLogin } from "@/hooks/use-api"
import { useAuthStore } from "@/stores/auth-store"
import { toast } from "sonner"

export function GoogleLoginButton() {
  const router = useRouter()
  const googleAuth = useGoogleLogin()

  return (
    <div className="flex justify-center [&_iframe]:rounded-full [&_div]:!bg-transparent [&_*]:!border-0">
      <GoogleLogin
        onSuccess={async (credentialResponse) => {
          const idToken = credentialResponse.credential
          if (!idToken) {
            toast.error("Failed to get Google credentials")
            return
          }

          try {
            const res = await googleAuth.mutateAsync(idToken)
            if (res.access_token && res.refresh_token && res.user) {
              useAuthStore.getState().setAuth(res.user, res.access_token, res.refresh_token)
              router.push("/dashboard")
            }
          } catch {
            toast.error("Google login failed")
          }
        }}
        onError={() => {
          toast.error("Google login failed")
        }}
        theme="filled_black"
        shape="pill"
        size="large"
        width={360}
        text="continue_with"
      />
    </div>
  )
}
