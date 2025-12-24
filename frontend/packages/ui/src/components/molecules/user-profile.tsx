import * as React from "react"
import { Avatar, AvatarImage, AvatarFallback } from "../atoms/avatar"
import { cn } from "../../lib/utils"

export interface UserProfileProps extends React.HTMLAttributes<HTMLDivElement> {
  name: string
  role: string
  avatarSrc?: string
}

const UserProfile = React.forwardRef<HTMLDivElement, UserProfileProps>(
  ({ name, role, avatarSrc, className, ...props }, ref) => {
    const initials = name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2)

    return (
      <div ref={ref} className={cn("flex items-center gap-3", className)} {...props}>
        <Avatar>
          {avatarSrc && <AvatarImage src={avatarSrc} alt={name} />}
          <AvatarFallback>{initials}</AvatarFallback>
        </Avatar>
        <div className="flex flex-col">
          <span className="text-sm font-medium text-foreground">{name}</span>
          <span className="text-xs text-muted-foreground">{role}</span>
        </div>
      </div>
    )
  }
)
UserProfile.displayName = "UserProfile"

export { UserProfile }
