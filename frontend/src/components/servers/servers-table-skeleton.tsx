import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

export function ServersTableSkeleton() {
  return (
    <Table>
      <TableHeader className="[&_tr]:border-b-primary-foreground/60">
        <TableRow className="hover:bg-background">
          <TableHead>
            <Skeleton className="h-4 w-4" />
          </TableHead>
          <TableHead>
            <div className="text-primary font-medium">Name</div>
          </TableHead>
          <TableHead>
            <div className="text-primary font-medium">Type</div>
          </TableHead>
          <TableHead>
            <div className="text-primary font-medium">Base URL</div>
          </TableHead>
          <TableHead>
            <div className="text-primary font-medium">Project</div>
          </TableHead>
          <TableHead>
            <div className="text-primary font-medium">Created At</div>
          </TableHead>
          <TableHead>
            <div className="text-primary font-medium">Actions</div>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, index) => (
          <TableRow key={index} className="border-b-primary-foreground">
            <TableCell>
              <Skeleton className="h-4 w-4" />
            </TableCell>
            <TableCell>
              <div className="flex items-center gap-2">
                <Skeleton className="h-5 w-5 rounded" />
                <Skeleton className="h-4 w-32" />
              </div>
            </TableCell>
            <TableCell>
              <Skeleton className="h-5 w-20 rounded-full" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-48" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-24" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-28" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-8 w-8 rounded-full" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

