export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterData {
  email: string
  password: string
  firstName: string
  lastName: string
  organizationName: string
  organizationSubdomain: string
}

export interface AuthError {
  message: string
  field?: string
}
