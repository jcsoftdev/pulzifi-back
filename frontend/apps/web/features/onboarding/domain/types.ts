export interface OnboardingFormValues {
  org_name: string
  subdomain: string
}

export interface OnboardingError {
  field?: 'org_name' | 'subdomain' | 'general'
  message: string
}
