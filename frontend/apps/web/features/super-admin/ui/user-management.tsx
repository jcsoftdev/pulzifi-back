'use client'

import { type AdminUser, SuperAdminApi } from '@workspace/services'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
} from '@workspace/ui/components/atoms'
import type { AxiosError } from 'axios'
import { ChevronLeft, ChevronRight, Loader2, ShieldCheck } from 'lucide-react'
import { useCallback, useEffect, useId, useRef, useState, useTransition } from 'react'
import { notification } from '@/lib/notification'

function getHttpStatus(err: unknown): number | undefined {
  return (err as AxiosError)?.response?.status
}

const PAGE_SIZE = 5

export function UserManagement() {
  const [isPending, startTransition] = useTransition()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loadError, setLoadError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const searchId = useId()

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const loadUsers = useCallback(async (currentSearch: string, currentPage: number) => {
    try {
      setLoadError(null)
      const data = await SuperAdminApi.listUsers({
        search: currentSearch,
        page: currentPage,
        pageSize: PAGE_SIZE,
      })
      setUsers(data.users ?? [])
      setTotal(data.total ?? 0)
    } catch (err) {
      const status = getHttpStatus(err)
      setLoadError(
        status === 403
          ? 'You need SUPER_ADMIN role to manage users.'
          : 'Failed to load users. Please try again.'
      )
    }
  }, [])

  // Mount-only fetch
  useEffect(() => {
    loadUsers('', 1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setSearch(value)
    setPage(1)

    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    debounceRef.current = setTimeout(() => {
      loadUsers(value, 1)
    }, 300)
  }

  const handlePageChange = (next: number) => {
    setPage(next)
    loadUsers(search, next)
  }

  const handlePromote = (user: AdminUser) => {
    setActionError(null)
    startTransition(async () => {
      try {
        await SuperAdminApi.promoteUser(user.id)
        notification.success({
          title: 'User promoted',
          description: `${user.email} is now a Super Admin.`,
        })
        await loadUsers(search, page)
      } catch (err) {
        const status = getHttpStatus(err)

        const userMessage =
          status === 409
            ? `${user.email} is already a Super Admin.`
            : status === 404
              ? 'User not found. They may have been deleted.'
              : 'Failed to promote user. Please try again.'

        setActionError(userMessage)
        notification.error({
          title: 'Promotion failed',
          description: userMessage,
        })
      }
    })
  }

  if (loadError) {
    return (
      <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
        <Card>
          <CardHeader>
            <CardTitle>Super Admin Access</CardTitle>
            <CardDescription>{loadError}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
      <Card>
        <CardHeader>
          <CardTitle>Promote users</CardTitle>
          <CardDescription>
            Search for a user by name or email and promote them to Super Admin.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          {actionError && (
            <div className="text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md px-4 py-3">
              {actionError}
            </div>
          )}

          <div className="flex items-center gap-3">
            <div className="flex-1">
              <label htmlFor={searchId} className="block text-sm font-medium text-foreground mb-2">
                Search by name or email
              </label>
              <Input
                id={searchId}
                placeholder="john@example.com"
                value={search}
                onChange={handleSearchChange}
                className="text-sm"
              />
            </div>
            {isPending && (
              <Loader2 className="w-5 h-5 animate-spin text-muted-foreground self-end mb-0.5" />
            )}
          </div>

          <div className="overflow-x-auto">
            {users.length === 0 ? (
              <div className="p-6 text-sm text-muted-foreground text-center">
                {search ? 'No users match your search.' : 'No users found.'}
              </div>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-3 px-4 font-medium">Name</th>
                    <th className="text-left py-3 px-4 font-medium">Email</th>
                    <th className="text-left py-3 px-4 font-medium">Role</th>
                    <th className="text-right py-3 px-4 font-medium">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id} className="border-b border-border hover:bg-muted/50">
                      <td className="py-3 px-4">
                        {user.firstName} {user.lastName}
                      </td>
                      <td className="py-3 px-4 text-muted-foreground">{user.email}</td>
                      <td className="py-3 px-4">
                        {user.isSuperAdmin && (
                          <Badge variant="default" className="gap-1">
                            <ShieldCheck className="w-3 h-3" />
                            Super Admin
                          </Badge>
                        )}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={isPending || user.isSuperAdmin}
                          onClick={() => handlePromote(user)}
                        >
                          {user.isSuperAdmin ? 'Already promoted' : 'Promote'}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>
                Page {page} of {totalPages} ({total} user{total !== 1 ? 's' : ''})
              </span>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1 || isPending}
                  onClick={() => handlePageChange(page - 1)}
                >
                  <ChevronLeft className="w-4 h-4" />
                  Prev
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages || isPending}
                  onClick={() => handlePageChange(page + 1)}
                >
                  Next
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
