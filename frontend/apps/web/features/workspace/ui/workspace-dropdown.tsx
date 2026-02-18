import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@workspace/ui/components/molecules/dropdown-menu'
import { ExternalLink, Pencil, Tag, Trash2 } from 'lucide-react'

export interface WorkspaceDropdownProps {
  trigger: React.ReactNode
  onOpen?: () => void
  onRename?: () => void
  onEditTag?: () => void
  onDelete?: () => void
}

export function WorkspaceDropdown({
  trigger,
  onOpen,
  onRename,
  onEditTag,
  onDelete,
}: Readonly<WorkspaceDropdownProps>) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{trigger}</DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-36 p-0.5 rounded-md">
        <DropdownMenuItem
          onClick={onOpen}
          className="flex items-center gap-1.5 px-1 py-1 rounded text-sm font-normal text-foreground cursor-pointer"
        >
          <ExternalLink className="w-3.5 h-3.5" />
          <span>Open</span>
        </DropdownMenuItem>

        <DropdownMenuItem
          onClick={onRename}
          className="flex items-center gap-1.5 px-1 py-1 rounded text-sm font-normal text-foreground cursor-pointer"
        >
          <Pencil className="w-3.5 h-3.5" />
          <span>Rename</span>
        </DropdownMenuItem>

        <DropdownMenuItem
          onClick={onEditTag}
          className="flex items-center gap-1.5 px-1 py-1 rounded text-sm font-normal text-foreground cursor-pointer"
        >
          <Tag className="w-3.5 h-3.5" />
          <span>Edit tag</span>
        </DropdownMenuItem>

        <DropdownMenuItem
          onClick={onDelete}
          className="flex items-center gap-1.5 px-1 py-1 rounded text-sm font-normal text-destructive cursor-pointer bg-background"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span>Delete</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
