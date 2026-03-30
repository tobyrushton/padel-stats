import { useMemo, useState } from "react"

import {
  ApiError,
  createApiClient,
  type ErrorResponse,
  type SigninInput,
} from "@/lib/api-client"
import { setAuthToken } from "@/lib/auth-token"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const SIGNIN_INITIAL_STATE: SigninInput = {
  username: "",
  password: "",
}

function getSignInErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    const body = error.body as ErrorResponse | undefined
    if (body?.error) {
      return body.error
    }

    if (error.status === 401) {
      return "Invalid username or password."
    }

    if (error.status === 400) {
      return "Please check your details and try again."
    }
  }

  return "Something went wrong. Please try again."
}

interface SignInFormProps {
  apiBaseUrl: string
}

export default function SignInForm({ apiBaseUrl }: SignInFormProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])
  const [form, setForm] = useState<SigninInput>(SIGNIN_INITIAL_STATE)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const handleChange = (field: keyof SigninInput, value: string) => {
    setForm((current) => ({
      ...current,
      [field]: value,
    }))
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setErrorMessage(null)

    if (!form.username || !form.password) {
      setErrorMessage("Username and password are required.")
      return
    }

    try {
      setIsSubmitting(true)
      const result = await apiClient.signIn(form)

      if (!result.token) {
        setErrorMessage("Authentication succeeded but no token was returned.")
        return
      }

      setAuthToken(result.token)
      window.location.assign("/")
    } catch (error) {
      setErrorMessage(getSignInErrorMessage(error))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-md space-y-5 rounded-xl border border-border bg-card p-6 text-card-foreground shadow-sm">
      <div className="space-y-1">
        <h1 className="text-xl font-semibold">Sign in</h1>
        <p className="text-sm text-muted-foreground">Welcome back. Enter your credentials to continue.</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="signin-username">Username</Label>
        <Input
          id="signin-username"
          autoComplete="username"
          value={form.username ?? ""}
          onChange={(event) => handleChange("username", event.currentTarget.value)}
          disabled={isSubmitting}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="signin-password">Password</Label>
        <Input
          id="signin-password"
          type="password"
          autoComplete="current-password"
          value={form.password ?? ""}
          onChange={(event) => handleChange("password", event.currentTarget.value)}
          disabled={isSubmitting}
          required
        />
      </div>

      {errorMessage ? (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {errorMessage}
        </p>
      ) : null}

      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? "Signing in..." : "Sign in"}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        Don&apos;t have an account?{" "}
        <a href="/auth/signup" className="font-medium text-foreground underline-offset-4 hover:underline">
          Create one
        </a>
      </p>
    </form>
  )
}
