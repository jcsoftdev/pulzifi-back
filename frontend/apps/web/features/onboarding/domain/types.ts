export interface OnboardingFormValues {
  org_name: string
  subdomain: string
}

export interface OnboardingError {
  field?: 'org_name' | 'subdomain' | 'general'
  message: string
}

export interface OnboardingProfileValues {
  company_size?: string
  business_type?: string
  competitor_challenges?: string[]
  website_url?: string
}
