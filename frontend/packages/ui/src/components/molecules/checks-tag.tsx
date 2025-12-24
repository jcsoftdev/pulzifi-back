import { forwardRef } from "react"

export interface ChecksTagProps {
  current: number
  max: number
  refillDate: string
}

const ChecksTag = forwardRef<HTMLDivElement, ChecksTagProps>(
  ({ current, max, refillDate }, ref) => {
    return (
      <div ref={ref} className="bg-accent border border-accent-foreground rounded-md px-3 py-1">
        <span className="text-[12.5px] font-normal text-foreground">
          {current}/{max} Checks  | Refill: {refillDate}
        </span>
      </div>
    )
  }
)
ChecksTag.displayName = "ChecksTag"

export { ChecksTag }
