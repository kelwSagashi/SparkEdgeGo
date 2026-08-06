import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { flexRender, getCoreRowModel, getFilteredRowModel, getPaginationRowModel, getSortedRowModel, useReactTable, type ColumnDef, type ColumnFiltersState, type SortingState, type VisibilityState } from "@tanstack/react-table";
import React from "react";
import { useInstancesStore } from '@/stores/instances-store';
import { Button } from "./ui/button";
import { ArrowUpDown, ChevronDown, MoreVertical, Play, Trash } from "lucide-react";
import { Checkbox } from "./ui/checkbox";
import { DropdownMenu, DropdownMenuCheckboxItem, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "./ui/dropdown-menu";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "./ui/alert-dialog";
import Search from "./search";
import { useNavigate } from "react-router-dom";
import type { Instance } from "@/rest-api-client/instances.service";
import { useShallow } from "zustand/react/shallow";

type InstanceType = Instance;

const HeaderColumn = ({
    title,
    column
}: {
    title: string, column: ColumnDef<InstanceType, unknown>
}) => {
    const [isHovered, setIsHovered] = React.useState(false);
    return (
        <div className="inline-flex items-center w-full justify-between"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
        >
            <div className="text-primary text-right font-medium">{title}</div>
            {isHovered && (
                <Button
                    variant={"ghost"}
                    className="rounded-full hover:bg-primary-foreground text-primary hover:text-primary/60"
                    // @ts-expect-error column typing
                    onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
                >
                    <ArrowUpDown />
                </Button>
            )}
        </div>
    )
}

const statusColors: Record<string, string> = {
    idle: 'bg-zinc-500',
    running: 'bg-emerald-500',
    paused: 'bg-amber-500',
    error: 'bg-red-500',
};

const columns: ColumnDef<InstanceType>[] = [
    {
        id: "select",
        header: ({ table }) => (
            <Checkbox
                checked={
                    table.getIsAllPageRowsSelected() ||
                    (table.getIsSomePageRowsSelected() && "indeterminate")
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
        accessorKey: "status",
        header: () => <></>,
        cell: ({ row }) => {
            const status = row.getValue("status") as string;
            return (
                <div className="p-2">
                    <div className={`w-3 h-3 rounded-full ${statusColors[status] || 'border border-primary-foreground'}`} />
                </div>
            )
        },
    },
    {
        accessorKey: "name",
        header: ({ column }) => <HeaderColumn title="Nome" column={column} />,
        cell: ({ row }) => <div className="capitalize">{row.getValue("name")}</div>,
    },
    {
        accessorKey: "id",
        header: ({ column }) => <HeaderColumn title="Id" column={column} />,
        cell: ({ row }) => <div className="font-mono text-xs">{(row.getValue("id") as string).slice(0, 12)}...</div>,
    },
    {
        accessorKey: "trigger_type",
        header: () => <div className="text-primary font-medium">Trigger</div>,
        cell: ({ row }) => <div className="capitalize">{row.getValue("trigger_type")}</div>,
    },
    {
        accessorKey: "created_at",
        header: () => <div className="text-primary font-medium">Criado em</div>,
        cell: ({ row }) => <div>{new Date(row.getValue("created_at")).toLocaleDateString('pt-BR')}</div>,
    },
    {
        accessorKey: "active",
        header: () => <div className="text-primary font-medium">Ativo</div>,
        cell: ({ row }) => <div className="capitalize">{row.getValue("active") ? 'Sim' : 'Não'}</div>,
    },
];


export default function InstanceTable() {
    const navigate = useNavigate();
    const [
        instances,
        fetchAll,
        triggerInstance,
        deleteInstance,
        loading
    ] = useInstancesStore(
        useShallow((s) => [
            s.instances,
            s.fetchAll,
            s.triggerInstance,
            s.deleteInstance,
            s.loading
        ])
    );
    const [sorting, setSorting] = React.useState<SortingState>([])
    const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([])
    const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({})
    const [rowSelection, setRowSelection] = React.useState({})

    const handleNavigate = (id: string) => {
        navigate(`/instances/${id}`);
    };

    React.useEffect(() => {
        fetchAll();
    }, [fetchAll]);

    const table = useReactTable({
        data: instances,
        columns,
        onSortingChange: setSorting,
        onColumnFiltersChange: setColumnFilters,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        getSortedRowModel: getSortedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
        onColumnVisibilityChange: setColumnVisibility,
        onRowSelectionChange: setRowSelection,
        state: { sorting, columnFilters, columnVisibility, rowSelection },
    })

    return (
        <section className="mt-4">
            <div className="flex flex-row justify-between items-stretch">
                <div className="flex gap-2">
                    <Search
                        value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
                        onChange={(event) => {
                            table.getColumn("name")?.setFilterValue(event.target.value)
                        }}
                    />
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="ml-auto text-primary hover:text-primary">
                                Colunas <ChevronDown />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="bg-background text-primary">
                            {table
                                .getAllColumns()
                                .filter((column) => column.getCanHide())
                                .map((column) => (
                                    <DropdownMenuCheckboxItem
                                        key={column.id}
                                        className="capitalize"
                                        checked={column.getIsVisible()}
                                        onCheckedChange={(value) => column.toggleVisibility(!!value)}
                                    >
                                        {column.id}
                                    </DropdownMenuCheckboxItem>
                                ))}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </div>
            <section className="mt-4">
                <div className="relative">
                    <div className="absolute overflow-x-hidden w-full">
                        <div className="flex flex-row">
                            <div className="w-full">
                                <Table>
                                    <TableHeader className="[&_tr]:border-b-primary-foreground/60">
                                        {table.getHeaderGroups().map((headerGroup) => (
                                            <TableRow className="hover:bg-background" key={headerGroup.id}>
                                                {headerGroup.headers.map((header) => (
                                                    <TableHead key={header.id}>
                                                        {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                                                    </TableHead>
                                                ))}
                                            </TableRow>
                                        ))}
                                    </TableHeader>
                                    <TableBody>
                                        {table.getRowModel().rows?.length ? (
                                            table.getRowModel().rows.map((row) => (
                                                <TableRow
                                                    onClick={() => handleNavigate(row.getValue("id"))}
                                                    key={row.id}
                                                    data-state={row.getIsSelected() && "selected"}
                                                    className="border-b-primary-foreground data-[state=selected]:bg-primary-foreground/40 hover:bg-secondary-foreground cursor-pointer"
                                                >
                                                    {row.getVisibleCells().map((cell) => (
                                                        <TableCell key={cell.id} className="h-14 text-primary">
                                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                                        </TableCell>
                                                    ))}
                                                </TableRow>
                                            ))
                                        ) : (
                                            <TableRow>
                                                <TableCell colSpan={columns.length} className="h-24 text-center">
                                                    Sem instâncias.
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
                                            <TableHead className="text-primary">Ações</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {table.getRowModel().rows?.length ? (
                                            table.getRowModel().rows.map((row) => (
                                                <TableRow key={row.id} className="border-b-primary-foreground hover:bg-secondary-foreground">
                                                    <TableCell className="h-14 hover:bg-transparent">
                                                        <Button disabled={loading} className="bg-transparent hover:bg-background/95 rounded-full" onClick={(e) => { e.stopPropagation(); triggerInstance(row.getValue("id")) }}>
                                                            <Play color="white" size={16} />
                                                        </Button>
                                                        <DropdownMenu>
                                                            <DropdownMenuTrigger asChild>
                                                                <Button className="bg-transparent hover:bg-background/95 rounded-full">
                                                                    <MoreVertical color="white" />
                                                                </Button>
                                                            </DropdownMenuTrigger>
                                                            <DropdownMenuContent align="start" side="left" className="bg-secondary-foreground">
                                                                <DropdownMenuItem
                                                                    className="text-primary focus:bg-primary-foreground"
                                                                    onClick={(e) => { e.stopPropagation(); navigator.clipboard.writeText(row.getValue("id")) }}
                                                                >
                                                                    Copiar ID
                                                                </DropdownMenuItem>
                                                                <DropdownMenuSeparator />
                                                                <DropdownMenuItem className="text-primary focus:bg-primary-foreground" onClick={(e) => { e.stopPropagation(); triggerInstance(row.getValue("id")) }} disabled={loading}>
                                                                    <Play color="white" size={14} />
                                                                    Executar
                                                                </DropdownMenuItem>
                                                            </DropdownMenuContent>
                                                        </DropdownMenu>
                                                        <AlertDialog>
                                                            <AlertDialogTrigger asChild>
                                                                <Button className="bg-transparent hover:bg-background/95 rounded-full" onClick={(e) => e.stopPropagation()}>
                                                                    <Trash color="white" />
                                                                </Button>
                                                            </AlertDialogTrigger>
                                                            <AlertDialogContent>
                                                                <AlertDialogHeader>
                                                                    <AlertDialogTitle>Você tem certeza?</AlertDialogTitle>
                                                                    <AlertDialogDescription>
                                                                        Essa ação não pode ser desfeita. Isso irá deletar
                                                                        permanentemente a instância.
                                                                    </AlertDialogDescription>
                                                                </AlertDialogHeader>
                                                                <AlertDialogFooter>
                                                                    <AlertDialogCancel>Cancelar</AlertDialogCancel>
                                                                    <AlertDialogAction onClick={(e) => { e.stopPropagation(); deleteInstance(row.getValue("id")) }} disabled={loading}>Continuar</AlertDialogAction>
                                                                </AlertDialogFooter>
                                                            </AlertDialogContent>
                                                        </AlertDialog>
                                                    </TableCell>
                                                </TableRow>
                                            ))
                                        ) : (
                                            <TableRow>
                                                <TableCell className="h-24 text-center">
                                                    Sem resultados.
                                                </TableCell>
                                            </TableRow>
                                        )}
                                    </TableBody>
                                </Table>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </section>
    )
}
