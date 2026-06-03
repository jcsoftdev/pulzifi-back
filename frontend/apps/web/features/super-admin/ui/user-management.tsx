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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms'
import { ChevronLeft, ChevronRight, Loader2, ShieldCheck } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useCallback, useEffect, useId, useRef, useState } from 'react'
import { UserDetailSheet } from './user-detail-sheet'

const PAGE_SIZE = 10

interface UserManagementProps {
  initialOrgId?: string
}

export function UserManagement({ initialOrgId = '' }: UserManagementProps) {
  const router = useRouter()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [orgId, setOrgId] = useState(initialOrgId)
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const searchId = useId()

  // Sync org filter when the parent propagates a new ?org= URL param
  useEffect(() => {
    setOrgId(initialOrgId)
    setPage(1)
  }, [initialOrgId])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const loadUsers = useCallback(async (s: string, p: number, org: string, st: string) => {
    setLoading(true)
    try {
      setLoadError(null)
      const data = await SuperAdminApi.listUsers({
        search: s,
        page: p,
        pageSize: PAGE_SIZE,
        orgId: org,
        status: st,
      })
      setUsers(data.users ?? [])
      setTotal(data.total ?? 0)
    } catch {
      setLoadError('You need SUPER_ADMIN role to manage users.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadUsers(search, page, orgId, status)
  }, [loadUsers, search, page, orgId, status])

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setPage(1)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => setSearch(value), 300)
  }

  const refresh = () => loadUsers(search, page, orgId, status)

  if (loadError) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Super Admin Access</CardTitle>
          <CardDescription>{loadError}</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Users</CardTitle>
          <CardDescription>
            Search, filter, and manage user accounts and their organization roles.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          <div className="flex flex-wrap items-end gap-3">
            <div className="flex-1 min-w-[220px]">
              <label htmlFor={searchId} className="block text-sm font-medium mb-2">
                Search by name or email
              </label>
              <Input id={searchId} placeholder="john@example.com" onChange={handleSearchChange} />
            </div>
            <Select
              value={status || 'all'}
              onValueChange={(v) => {
                setStatus(v === 'all' ? '' : v)
                setPage(1)
              }}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All statuses</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="suspended">Suspended</SelectItem>
                <SelectItem value="trial_expired">Trial expired</SelectItem>
              </SelectContent>
            </Select>
            {loading && <Loader2 className="w-5 h-5 animate-spin text-muted-foreground mb-2" />}
          </div>

          {orgId && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>Filtered by organization.</span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setOrgId('')
                  setPage(1)
                  router.replace('/admin?tab=users')
                }}
              >
                Clear org filter
              </Button>
            </div>
          )}

          <div className="overflow-x-auto">
            {users.length === 0 ? (
              <div className="p-6 text-sm text-muted-foreground text-center">No users found.</div>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-3 px-4 font-medium">Name</th>
                    <th className="text-left py-3 px-4 font-medium">Email</th>
                    <th className="text-left py-3 px-4 font-medium">Status</th>
                    <th className="text-left py-3 px-4 font-medium">Orgs</th>
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
                        <Badge variant={user.status === 'suspended' ? 'destructive' : 'secondary'}>
                          {user.status}
                        </Badge>
                      </td>
                      <td className="py-3 px-4">{user.orgCount}</td>
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
                          onClick={() => setSelectedId(user.id)}
                        >
                          Manage
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
                  disabled={page <= 1 || loading}
                  onClick={() => setPage((p) => p - 1)}
                >
                  <ChevronLeft className="w-4 h-4" /> Prev
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages || loading}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {selectedId && (
        <UserDetailSheet
          userId={selectedId}
          open={!!selectedId}
          onClose={() => setSelectedId(null)}
          onChanged={refresh}
        />
      )}
    </>
  )
}
