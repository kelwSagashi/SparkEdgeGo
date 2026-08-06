import React, { useState } from 'react';
import { Input } from '../ui/input';
import { Box, ChevronRight, Quote, Plus, Trash2, List, MoveDown, HelpCircle, GripVertical, Check } from 'lucide-react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../ui/collapsible';
import { cn } from '@/lib/utils';
import DraggableValue from '../DraggableValue';
import { ItemTypes } from '@/lib/constants';
import { Button } from '../ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { useDrop } from 'react-dnd';

export type JsonFieldType = 'string' | 'number' | 'boolean' | 'object' | 'array';

export function JsonViewMain({
    className,
    mainClassName,
    data,
    onParamChange,
    onAddField,
    onDeleteField,
    onDestructure,
    inputProps,
    pProps,
    defaultExpandLevel = 2,
    icons,
    forceExpand = false,
    expandOnce = true,
    filter,
    draggableValue = true,
    rootPath = '',
    ...props
}: React.ComponentProps<"div"> & {
    mainClassName?: string,
    data: any,
    onParamChange?: (path: string, key: string, value: any) => void;
    onAddField?: (path: string, key: string, type: JsonFieldType) => void;
    onDeleteField?: (path: string, key: string) => void;
    onDestructure?: (path: string, key: string, value: any) => void;
    inputProps?: React.ComponentProps<"input">;
    pProps?: React.ComponentProps<"p">;
    defaultExpandLevel?: number;
    filter?: string,
    forceExpand?: boolean,
    expandOnce?: boolean,
    icons?: {
        object?: React.ReactNode;
        string?: React.ReactNode;
        number?: React.ReactNode;
        boolean?: React.ReactNode;
    };
    draggableValue?: boolean;
    rootPath?: string;
}) {
    const [_expandOnce, setExpandOnce] = React.useState(expandOnce)
    const isArray = Array.isArray(data);

    return (
        <div className={cn('flex flex-col')}>
            <div className='flex min-h-0 flex-1 flex-col gap-2 overflow-auto group-data-[collapsible=icon]:overflow-hidden'>
                <div className='relative flex w-full min-w-0 flex-col p-2'>
                    <div className='w-full text-sm'>
                        <div className={cn('flex w-full min-w-0 flex-col gap-1', className)} {...props}>
                            {data && Object.entries(data).map(([name, value]) => (
                                <JsonView
                                    key={name}
                                    name={name}
                                    value={value}
                                    path={rootPath ? `${rootPath}.${name}` : name}
                                    level={0}
                                    expandOnce={!_expandOnce}
                                    setExpandOnce={setExpandOnce}
                                    defaultExpandLevel={defaultExpandLevel}
                                    forceExpand={forceExpand}
                                    icons={icons}
                                    onParamChange={onParamChange}
                                    onAddField={onAddField}
                                    onDeleteField={onDeleteField}
                                    onDestructure={onDestructure}
                                    inputProps={inputProps}
                                    pProps={pProps}
                                    draggableValue={draggableValue}
                                />
                            ))}
                            {onAddField && (
                                <AddFieldRow data={data} isArray={isArray} onAdd={(key, type) => onAddField('', key, type)} />
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

function AddFieldRow({ onAdd, isArray, data }: { onAdd: (key: string, type: JsonFieldType) => void, isArray?: boolean, data?: any }) {
    const [key, setKey] = useState('');
    const [type, setType] = useState<JsonFieldType>('string');

    const handleAdd = () => {
        if (!key && !isArray && data !== null && data !== undefined) return;
        onAdd(key, type);
        setKey('');
    };

    return (
        <div className="flex gap-1 mt-1 pl-1">
            {!isArray && <Input 
                placeholder="Nova chave..." 
                value={key}
                onChange={(e) => setKey(e.target.value)}
                className="h-9 text-[10px] bg-input border-border w-24 text-primary placeholder:text-secondary"
            />}
            <Select value={type} onValueChange={(v) => setType(v as any)}>
                <SelectTrigger className="h-9 text-[10px] bg-input border-border w-full px-2 text-primary">
                    <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-border">
                    <SelectItem value="string" className="text-zinc-300">Texto</SelectItem>
                    <SelectItem value="number" className="text-zinc-300">Número</SelectItem>
                    <SelectItem value="boolean" className="text-zinc-300">Booleano</SelectItem>
                    <SelectItem value="object" className="text-zinc-300">Objeto</SelectItem>
                    <SelectItem value="array" className="text-zinc-300">Lista</SelectItem>
                </SelectContent>
            </Select>
            <Button 
                type="button"
                size="icon"
                onClick={handleAdd}
                className="h-9 w-9 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 border-none"
            >
                <Check className="w-9 h-9" />
            </Button>
        </div>
    );
}


export function JsonButtonView({
    className,
    ...props
}: React.ComponentProps<"button">) {
    return (
        <button className={
            cn('flex items-center gap-2 overflow-hidden',
                'rounded-md p-2 text-left text-sm outline-hidden ring-sidebar-ring',
                'transition-[width,height,padding]',
                'focus-visible:ring-2',
                'disabled:pointer-events-none disabled:opacity-50',
                'pointer-events-auto cursor-pointer aria-disabled:opacity-50',
                'data-[active=true]:font-medium',
                'group-data-[collapsible=icon]:size-8!',
                'group-data-[collapsible=icon]:p-2! [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0',
                'text-primary',
                className
            )}
            type="button"
            {...props} />
    )
}

export function JsonView({
    name,
    value,
    path,
    level,
    defaultExpandLevel,
    forceExpand,
    icons,
    inputProps,
    pProps,
    expandOnce,
    setExpandOnce,
    onParamChange,
    onAddField,
    onDeleteField,
    onDestructure,
    draggableValue = true
}: {
    name: string;
    value: any;
    path: string;
    level: number;
    defaultExpandLevel: number;
    forceExpand: boolean | null;
    expandOnce: boolean,
    setExpandOnce: React.Dispatch<React.SetStateAction<boolean>>,
    icons?: Record<string, React.ReactNode>;
    onParamChange?: (path: string, key: string, value: any) => void;
    onAddField?: (path: string, key: string, type: JsonFieldType) => void;
    onDeleteField?: (path: string, key: string) => void;
    onDestructure?: (path: string, key: string, value: any) => void;
    inputProps?: React.ComponentProps<"input">;
    pProps?: React.ComponentProps<"p">;
    draggableValue?: boolean
}) {
    const isObject = typeof value === "object" && value !== null;
    const [open, setOpen] = React.useState(forceExpand ? forceExpand : level < defaultExpandLevel);

    const [{ isOver }, drop] = useDrop(() => ({
        accept: ItemTypes.OUTPUT_VALUE,
        drop: (item: { value: string, data?: any }) => {
            if (onParamChange) {
                const parentPath = path.split('.').slice(0, -1).join('.');
                if (isObject) {
                    if (item.data && typeof item.data === 'object') {
                        onParamChange(parentPath, name, item.data);
                    } else {
                        onParamChange(parentPath, name, `{{${item.value}}}`);
                    }
                } else {
                    if (item.data && typeof item.data === 'object' && typeof value === 'string') {
                        onParamChange(parentPath, name, JSON.stringify(item.data, null, 2));
                    } else {
                        onParamChange(parentPath, name, `{{${item.value}}}`);
                    }
                }
            }
        },
        collect: (monitor) => ({
            isOver: !!monitor.isOver(),
        }),
    }), [name, path, onParamChange, isObject, value]);

    const handleSetExpand = React.useCallback((value: boolean) => {
        setExpandOnce(false);
        setOpen(value);
    }, [expandOnce]);

    const icon =
        Array.isArray(value)
            ? <List className="w-4 h-4" />
            : typeof value === "object" && value !== null
                ? icons?.object ?? <Box className="w-4 h-4" />
                : typeof value === "string"
                    ? icons?.string ?? <Quote className="w-4 h-4" />
                    : typeof value === "number"
                        ? icons?.number ?? <span className="text-blue-400 font-bold">#</span>
                        : typeof value === "boolean"
                            ? icons?.boolean ?? <span className="text-amber-400">⚙</span>
                            : <HelpCircle className="w-4 h-4 text-zinc-600" />;

    if (!isObject) {
        return (
            <div 
                ref={onParamChange ? drop as any : undefined}
                className={cn(
                    "flex items-center group/item rounded transition-colors",
                    isOver && "bg-violet-500/20 ring-1 ring-violet-500/50"
                )}
            >
                <div className="flex-1 flex items-center min-w-0">
                    <JsonButtonView className="flex-1 data-[active=true]:bg-transparent pr-1">
                        <div className='w-full flex gap-2 items-center'>
                            <span className='bg-zinc-800/80 p-1.5 rounded flex flex-row items-center justify-center gap-1 border border-white/[0.05] shrink-0'>
                                {icon}
                                <p className="text-[10px] text-zinc-400 font-mono">{name}</p>
                            </span>
                            
                            {onParamChange ? (
                                <div className="flex-1 min-w-0 flex items-center gap-2">
                                     <GripVertical className="w-3 h-3 text-zinc-600 shrink-0 opacity-0 group-hover/item:opacity-100" />
                                     <Input
                                        value={value === null || value === undefined ? "" : value}
                                        onChange={(e) => onParamChange(path.split('.').slice(0, -1).join('.'), name, e.target.value)}
                                        className="h-7 text-[10px] py-1 bg-input border-border text-primary placeholder:text-secondary font-mono focus:ring-violet-500/30"
                                        placeholder="Valor ou {{path}}"
                                        {...inputProps}
                                    />
                                </div>
                            ) : (
                                <div className="flex flex-col min-w-0">
                                    {draggableValue ? (
                                        <DraggableValue type={ItemTypes.OUTPUT_VALUE} isDropped={false} value={path} data={value}>
                                            <p className="text-emerald-400 font-mono text-[10px] truncate max-w-[150px]">{String(value)}</p>
                                        </DraggableValue>
                                    ) : (
                                        <div className={cn(
                                            "px-1.5 py-0.5 rounded text-[8px] uppercase tracking-wider font-bold w-fit",
                                            typeof value === 'string' && "bg-blue-500/10 text-blue-400 border border-blue-500/20",
                                            typeof value === 'number' && "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20",
                                            typeof value === 'boolean' && "bg-amber-500/10 text-amber-400 border border-amber-500/20",
                                            (value === null || value === undefined) && "bg-zinc-800 text-zinc-500 border border-white/5"
                                        )}>
                                            {String(value)}
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    </JsonButtonView>
                    
                    <div className="flex items-center gap-0.5 px-1">
                        {onDeleteField && (
                            <Button 
                                type="button" 
                                variant="ghost" 
                                size="icon" 
                                onClick={() => onDeleteField(path.split('.').slice(0, -1).join('.'), name)}
                                className="h-6 w-6 text-zinc-600 hover:text-red-400 opacity-0 group-hover/item:opacity-100 transition-opacity"
                            >
                                <Trash2 className="w-3 h-3" />
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className='relative'>
            <Collapsible
                open={open}
                onOpenChange={handleSetExpand}
                className="group/collapsible [&[data-state=open]>div>button>svg:first-child]:rotate-90"
            >
                <div 
                    ref={onParamChange ? drop as any : undefined}
                    className={cn(
                        "flex items-center group/item rounded transition-colors",
                         isOver && "bg-violet-500/20 ring-1 ring-violet-500/50"
                    )}
                >
                    <CollapsibleTrigger asChild>
                        <JsonButtonView className="flex-1 flex flex-row items-center gap-1">
                            <ChevronRight className="transition-transform w-3 h-3 text-zinc-600 shrink-0" />
                            {draggableValue ? (
                                <DraggableValue type={ItemTypes.OUTPUT_VALUE} isDropped={false} value={path} data={value}>
                                    <div className="flex items-center gap-1.5 w-full">
                                        {icon}
                                        <span className="text-zinc-400 font-medium text-[11px]">{name}</span>
                                    </div>
                                </DraggableValue>
                            ) : (
                                <div className="flex items-center gap-1.5 w-full">
                                    {icon}
                                    <span className="text-zinc-400 font-medium text-[11px]">{name}</span>
                                </div>
                            )}
                        </JsonButtonView>
                    </CollapsibleTrigger>
                    
                    <div className="flex items-center gap-0.5 px-1 opacity-0 group-hover/item:opacity-100 transition-opacity">
                        {onDestructure && isObject && !Array.isArray(value) && (
                            <Button 
                                type="button" 
                                variant="ghost" 
                                size="icon" 
                                title="Desestruturar Objeto"
                                onClick={() => onDestructure(path.split('.').slice(0, -1).join('.'), name, value)}
                                className="h-6 w-6 text-zinc-600 hover:text-violet-400"
                            >
                                <MoveDown className="w-3 h-3" />
                            </Button>
                        )}
                        {onDeleteField && (
                            <Button 
                                type="button" 
                                variant="ghost" 
                                size="icon" 
                                onClick={() => onDeleteField(path.split('.').slice(0, -1).join('.'), name)}
                                className="h-6 w-6 text-zinc-600 hover:text-red-400"
                            >
                                <Trash2 className="w-3 h-3" />
                            </Button>
                        )}
                    </div>
                </div>
                <CollapsibleContent>
                    <div
                        className={cn(
                            "border-white/[0.06] mx-3.5 flex min-w-0 translate-x-px",
                            "flex-col gap-1 border-l px-2.5 py-0.5",
                        )}
                    >
                        {Object.entries(value).map(([childName, childValue]) => {
                            const childPath = `${path}.${childName}`;
                            return (
                                <JsonView
                                    key={childName}
                                    name={childName}
                                    value={childValue}
                                    path={childPath}
                                    level={level + 1}
                                    defaultExpandLevel={expandOnce ? 0 : defaultExpandLevel}
                                    forceExpand={forceExpand && expandOnce}
                                    icons={icons}
                                    onParamChange={onParamChange}
                                    onAddField={onAddField}
                                    onDeleteField={onDeleteField}
                                    onDestructure={onDestructure}
                                    expandOnce={expandOnce}
                                    setExpandOnce={setExpandOnce}
                                    inputProps={inputProps}
                                    pProps={pProps}
                                />
                            );
                        })}
                        {onAddField && (
                            <AddFieldRow data={value} isArray={Array.isArray(value)} onAdd={(key, type) => onAddField(path, key, type)} />
                        )}
                    </div>
                </CollapsibleContent>
            </Collapsible>
        </div>
    )
}


