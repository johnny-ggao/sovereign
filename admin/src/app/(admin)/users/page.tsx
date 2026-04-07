'use client'

import { useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import {
  NavBar,
  SearchBar,
  Tag,
  Skeleton,
  PullRefresh,
  LoadMore,
} from '@arco-design/mobile-react'
import type { LoadMoreRef } from '@arco-design/mobile-react/esm/load-more'
import { useUserList } from '@/hooks/use-api'
import type { UserListItem } from '@/types/api'

function UserCard({
  user,
  onClick,
}: {
  user: UserListItem
  onClick: () => void
}) {
  return (
    <div
      onClick={onClick}
      style={{
        background: '#fff',
        borderRadius: 12,
        padding: '14px 16px',
        cursor: 'pointer',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: '#1d2129',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {user.full_name || 'Unnamed'}
          </div>
          <div
            style={{
              fontSize: 12,
              color: '#86909c',
              marginTop: 4,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {user.email}
          </div>
        </div>
        <div style={{ textAlign: 'right', marginLeft: 12 }}>
          <div style={{ fontSize: 15, fontWeight: 600, color: '#1d2129' }}>
            ${user.balance}
          </div>
          <Tag
            style={{
              marginTop: 4,
              fontSize: 10,
              borderColor: user.is_active ? '#00b42a' : '#f53f3f',
              color: user.is_active ? '#00b42a' : '#f53f3f',
            }}
          >
            {user.is_active ? 'Active' : 'Disabled'}
          </Tag>
        </div>
      </div>
    </div>
  )
}

export default function UsersPage() {
  const router = useRouter()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [allUsers, setAllUsers] = useState<UserListItem[]>([])
  const [hasMore, setHasMore] = useState(true)
  const loadMoreRef = useRef<LoadMoreRef>(null)
  const limit = 20

  const { data, isLoading, refetch } = useUserList({
    page,
    limit,
    search: search || undefined,
  })

  const handleRefresh = useCallback(async () => {
    setPage(1)
    setAllUsers([])
    setHasMore(true)
    await refetch()
  }, [refetch])

  const handleSearchChange = useCallback(
    (_e: React.ChangeEvent<HTMLInputElement>, val: string) => {
      setSearch(val)
      setPage(1)
      setAllUsers([])
      setHasMore(true)
    },
    []
  )

  const displayUsers =
    page === 1 && data ? data.data : [...allUsers, ...(data?.data ?? [])]

  return (
    <div>
      <NavBar title="Users" leftContent={null} />
      <div style={{ padding: '0 16px 16px' }}>
        <SearchBar
          placeholder="Search by name or email"
          onChange={handleSearchChange}
          style={{ marginBottom: 12 }}
        />
        <PullRefresh onRefresh={handleRefresh}>
          <div style={{ minHeight: 200 }}>
            {isLoading && page === 1 ? (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 12,
                }}
              >
                <Skeleton animation="gradient" />
                <Skeleton animation="gradient" />
                <Skeleton animation="gradient" />
              </div>
            ) : (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 8,
                }}
              >
                {displayUsers.length === 0 && (
                  <div
                    style={{
                      textAlign: 'center',
                      padding: 40,
                      color: '#86909c',
                      fontSize: 14,
                    }}
                  >
                    No users found
                  </div>
                )}
                {displayUsers.map((user) => (
                  <UserCard
                    key={user.id}
                    user={user}
                    onClick={() => router.push(`/users/${user.id}`)}
                  />
                ))}
              </div>
            )}
            {hasMore && displayUsers.length > 0 && (
              <LoadMore
                ref={loadMoreRef}
                getData={(callback) => {
                  const nextPage = page + 1
                  setPage(nextPage)
                  setAllUsers(displayUsers)
                  const total = data?.meta?.total ?? 0
                  if (displayUsers.length >= total) {
                    setHasMore(false)
                    callback('nomore')
                  } else {
                    callback('prepare')
                  }
                }}
                getDataAtFirst={false}
              />
            )}
          </div>
        </PullRefresh>
      </div>
    </div>
  )
}
