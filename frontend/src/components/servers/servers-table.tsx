import React from 'react';
import { format } from 'date-fns';
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { MoreVertical, Server, Database, Cloud, Edit, Eye, Trash } from 'lucide-react';
import type { ServerWithRelations } from '@/interfaces/servers';
import { Badge } from '../ui/badge';

interface ServersTableProps {
  servers: ServerWithRelations[];
  onEdit?: (server: ServerWithRelations) => void;
  onViewEndpoints?: (server: ServerWithRelations) => void;
  onDelete: (id: string) => void;
  loading?: boolean;
}

// Icon mapping for server types
const getServerTypeIcon = (typeKey?: string) => {
  switch (typeKey) {
    case 'database':
      return <Database className="h-4 w-4" />;
    case 'google_drive':
    case 'cloud':
      return <Cloud className="h-4 w-4" />;
    default:
      return <Server className="h-4 w-4" />;
  }
};

export function ServersTable({
  servers,
  onEdit,
  onViewEndpoints,
  onDelete,
  loading = false,
}: ServersTableProps) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  const columns: ColumnDef<ServerWithRelations>[] = [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && 'indeterminate')
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      header: () => <div className="text-primary font-medium">Name</div>,
      cell: ({ row }) => {
        const server = row.original;
        return (
          <div className="flex items-center gap-2">
            <div className="text-primary">
              {getServerTypeIcon(server.serverType?.key)}
            </div>
            <span className="font-semibold text-primary">{server.name}</span>
          </div>
        );
      },
    },
    {
      accessorKey: 'type',
      header: () => <div className="text-primary font-medium">Type</div>,
      cell: ({ row }) => {
        const server = row.original;
        return (
          <Badge variant="secondary" className="bg-primary-foreground/40 text-primary">
            {server.serverType?.name || 'Unknown'}
          </Badge>
        );
      },
    },
    {
      accessorKey: 'base_url',
      header: () => <div className="text-primary font-medium">Base URL</div>,
      cell: ({ row }) => (
        <div className="text-primary max-w-md truncate" title={row.getValue('base_url')}>
          {row.getValue('base_url')}
        </div>
      ),
    },
    {
      accessorKey: 'project_id',
      header: () => <div className="text-primary font-medium">Project</div>,
      cell: ({ row }) => {
        const server = row.original;
        return (
          <div className="text-primary">
            {server.project?.name || 'N/A'}
          </div>
        );
      },
    },
    {
      accessorKey: 'created_at',
      header: () => <div className="text-primary font-medium">Created At</div>,
      cell: ({ row }) => {
        const date = row.getValue('created_at') as string;
        return (
          <div className="text-primary">
            {format(new Date(date), 'MMM dd, yyyy')}
          </div>
        );
      },
    },
  ];

  const table = useReactTable({
    data: servers,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div className="relative">
      <div className="flex flex-row">
        <div className="w-full">
          <Table>
            <TableHeader className="[&_tr]:border-b-primary-foreground/60">
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow className="hover:bg-background" key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className="border-b-primary-foreground data-[state=selected]:bg-primary-foreground/40 hover:bg-secondary-foreground"
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id} className="h-14 text-primary">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={columns.length}
                    className="h-24 text-center text-primary"
                  >
                    No results.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
        <div>
          <Table className="bg-primary-foreground/20">
            <TableHeader className="[&_tr]:border-b-primary-foreground/60">
              <TableRow>
                <TableHead className="text-primary">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className="border-b-primary-foreground data-[state=selected]:bg-primary-foreground/40 hover:bg-secondary-foreground"
                  >
                    <TableCell className="h-14">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            className="bg-transparent hover:bg-background/95 rounded-full"
                            disabled={loading}
                          >
                            <MoreVertical color="white" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent
                          align="start"
                          side="left"
                          className="bg-secondary-foreground"
                        >
                          <DropdownMenuItem
                            className="text-primary focus:bg-primary-foreground"
                            onClick={(e) => {
                              e.stopPropagation();
                              navigator.clipboard.writeText(row.original.id);
                            }}
                          >
                            Copy ID
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          {onEdit && (
                            <DropdownMenuItem
                              className="text-primary focus:bg-primary-foreground"
                              onClick={(e) => {
                                e.stopPropagation();
                                onEdit(row.original);
                              }}
                            >
                              <Edit className="mr-2 h-4 w-4" />
                              Edit
                            </DropdownMenuItem>
                          )}
                          {onViewEndpoints && (
                            <DropdownMenuItem
                              className="text-primary focus:bg-primary-foreground"
                              onClick={(e) => {
                                e.stopPropagation();
                                onViewEndpoints(row.original);
                              }}
                            >
                              <Eye className="mr-2 h-4 w-4" />
                              View Endpoints
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <DropdownMenuItem
                                className="text-red-500 focus:bg-primary-foreground focus:text-red-500"
                                onSelect={(e) => e.preventDefault()}
                              >
                                <Trash className="mr-2 h-4 w-4" />
                                Delete
                              </DropdownMenuItem>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogHeader>
                                <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                                <AlertDialogDescription>
                                  This action cannot be undone. This will permanently delete
                                  the server "{row.original.name}" and all its endpoints.
                                </AlertDialogDescription>
                              </AlertDialogHeader>
                              <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    onDelete(row.original.id);
                                  }}
                                  disabled={loading}
                                  className="bg-red-500 hover:bg-red-600"
                                >
                                  Delete
                                </AlertDialogAction>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell className="h-24 text-center text-primary">
                    No results.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}

