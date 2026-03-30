import { useMemo, useState } from "react"

import {
  ApiError,
  createApiClient,
  type ErrorResponse,
  type SignupInput,
} from "@/lib/api-client"
import { setAuthToken } from "@/lib/auth-token"
import { setAuthUser } from "@/lib/auth-user"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const SIGNUP_INITIAL_STATE: SignupInput = {
  firstName: "",
  lastName: "",
  username: "",
  password: "",
}

function getSignUpErrorMessage(error: unknown): string {
    console.log("Error during sign up:", error)
  if (error instanceof ApiError) {
    const body = error.body as ErrorResponse | undefined
    if (body?.error) {
      return body.error
    }

    if (error.status === 409) {
      return "That username is already taken."
    }

    if (error.status === 400) {
      return "Please review the form values and try again."
    }
  }

  return "Something went wrong. Please try again."
}

interface SignUpFormProps {
  apiBaseUrl: string
}

export default function SignUpForm({ apiBaseUrl }: SignUpFormProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])
  const [form, setForm] = useState<SignupInput>(SIGNUP_INITIAL_STATE)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const handleChange = (field: keyof SignupInput, value: string) => {
    setForm((current) => ({
      ...current,
      [field]: value,
    }))
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setErrorMessage(null)

    if (!form.firstName || !form.lastName || !form.username || !form.password) {
      setErrorMessage("All fields are required.")
      return
    }

    try {
      setIsSubmitting(true)
      const result = await apiClient.signUp(form)

      if (!result.token) {
        setErrorMessage("Account created but no token was returned.")
        return
      }

      setAuthToken(result.token)
      setAuthUser(result.user)
      window.location.assign("/")
    } catch (error) {
      setErrorMessage(getSignUpErrorMessage(error))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-md space-y-5 rounded-xl border border-border bg-card p-6 text-card-foreground shadow-sm">
      <div className="space-y-1">
        <h1 className="text-xl font-semibold">Create account</h1>
        <p className="text-sm text-muted-foreground">Set up your account to start tracking your padel stats.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="signup-first-name">First name</Label>
          <Input
            id="signup-first-name"
            autoComplete="given-name"
            value={form.firstName ?? ""}
            onChange={(event) => handleChange("firstName", event.currentTarget.value)}
            disabled={isSubmitting}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="signup-last-name">Last name</Label>
          <Input
            id="signup-last-name"
            autoComplete="family-name"
            value={form.lastName ?? ""}
            onChange={(event) => handleChange("lastName", event.currentTarget.value)}
            disabled={isSubmitting}
            required
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="signup-username">Username</Label>
        <Input
          id="signup-username"
          autoComplete="username"
          value={form.username ?? ""}
          onChange={(event) => handleChange("username", event.currentTarget.value)}
          disabled={isSubmitting}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="signup-password">Password</Label>
        <Input
          id="signup-password"
          type="password"
          autoComplete="new-password"
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
        {isSubmitting ? "Creating account..." : "Create account"}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <a href="/auth/signin" className="font-medium text-foreground underline-offset-4 hover:underline">
          Sign in
        </a>
      </p>
    </form>
  )
}
