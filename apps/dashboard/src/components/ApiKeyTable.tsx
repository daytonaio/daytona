import { ApiKeyList } from '@daytonaio/api-client'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { TableHeader, TableRow, TableHead, TableBody, TableCell, Table } from './ui/table'
import { Button } from './ui/button'
import { useState } from 'react'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { Pagination } from './Pagination'
import { Loader2 } from 'lucide-react'

interface DataTableProps {
  data: ApiKeyList[]
  loading: boolean
  loadingKeys: Record<string, boolean>
  onRevoke: (keyName: string) => void
}

export function ApiKeyTable({ data, loading, loadingKeys, onRevoke }: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const columns = getColumns({ onRevoke, loadingKeys })
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting,
    },
  })

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={`${loadingKeys[row.original.name] ? 'opacity-50 pointer-events-none' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              !loading && (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    No results.
                  </TableCell>
                </TableRow>
              )
            )}
          </TableBody>
        </Table>
      </div>
      <Pagination table={table} className="mt-4" />
    </div>
  )
}

const getColumns = ({
  onRevoke,
  loadingKeys,
}: {
  onRevoke: (keyName: string) => void
  loadingKeys: Record<string, boolean>
}): ColumnDef<ApiKeyList>[] => {
  const columns: ColumnDef<ApiKeyList>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
    },
    {
      accessorKey: 'value',
      header: 'Key',
    },
    {
      accessorKey: 'permissions',
      header: () => {
        return <div className="max-w-md px-3">Permissions</div>
      },
      cell: ({ row }) => {
        const permissions = row.original.permissions.join(', ')
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <div className="truncate max-w-md px-3 cursor-text">{permissions || '-'}</div>
              </TooltipTrigger>
              {permissions && (
                <TooltipContent>
                  <p className="max-w-[300px]">{permissions}</p>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        )
      },
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      cell: ({ row }) => {
        return new Date(row.original.createdAt).toLocaleDateString()
      },
    },
    {
      id: 'actions',
      header: () => {
        return <div className="px-4">Actions</div>
      },
      cell: ({ row }) => {
        const isLoading = loadingKeys[row.original.name]

        return (
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" disabled={isLoading} className="w-20" title="Revoke Key">
                {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Revoke'}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Confirm Key Revocation</DialogTitle>
                <DialogDescription>
                  Are you absolutely sure? This action cannot be undone. This will permanently delete this API key.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button type="button" variant="secondary">
                    Close
                  </Button>
                </DialogClose>
                <DialogClose asChild>
                  <Button variant="destructive" onClick={() => onRevoke(row.original.name)}>
                    Revoke
                  </Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )
      },
    },
  ]

  return columns
}
