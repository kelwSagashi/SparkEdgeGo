import React, { useMemo, useRef, useState } from 'react';
import { Input } from '../ui/input';
import { Box, ChevronRight, Quote, Plus, Trash2, List, MoveDown, HelpCircle, GripVertical, Check } from 'lucide-react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../ui/collapsible';
import { cn } from '@/lib/utils';
import DraggableValue from '../DraggableValue';
import { ItemTypes } from '@/lib/constants';
import { Button } from '../ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { useDrop } from 'react-dnd';
import { Popover, PopoverAnchor, PopoverContent } from '../ui/popover';
import { Command, CommandEmpty, CommandGroup, CommandItem, CommandList } from '../ui/command';

export type JsonFieldType = 'string' | 'number' | 'boolean' | 'object' | 'array';
type TemplateSuggestion = {
    path: string;
    label: string;
    kind: 'object' | 'array' | 'value';
};

function isPlainObject(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function buildTemplateSuggestions(
    value: unknown,
    prefix = '$',
    label = '$',
): TemplateSuggestion[] {
    if (Array.isArray(value)) {
        const current: TemplateSuggestion[] = [{ path: prefix, label, kind: 'array' }];
        value.forEach((item, index) => {
            current.push(...buildTemplateSuggestions(item, `${prefix}.${index}`, `${label}.${index}`));
        });
        return current;
    }

    if (isPlainObject(value)) {
        const current: TemplateSuggestion[] = [{ path: prefix, label, kind: 'object' }];
        Object.entries(value).forEach(([key, nested]) => {
            current.push(...buildTemplateSuggestions(nested, `${prefix}.${key}`, `${label}.${key}`));
        });
        return current;
    }

    return [{ path: prefix, label, kind: 'value' }];
}

function getTemplateContext(value: string, cursor: number) {
    const beforeCursor = value.slice(0, cursor);
    const lastOpen = beforeCursor.lastIndexOf('{{');
    if (lastOpen !== -1) {
        const closeAfterOpen = value.indexOf('}}', lastOpen + 2);
        if (closeAfterOpen === -1 || closeAfterOpen >= cursor) {
            const rawQuery = value.slice(lastOpen + 2, cursor);
            return {
                mode: 'placeholder' as const,
                start: lastOpen,
                end: closeAfterOpen === -1 ? cursor : closeAfterOpen + 2,
                query: rawQuery.trim(),
            };
        }
    }

    let tokenStart = cursor;
    while (tokenStart > 0) {
        const previous = value[tokenStart - 1];
        if (/\s|["',:[\]{}()]/.test(previous)) {
            break;
        }
        tokenStart -= 1;
    }

    const token = value.slice(tokenStart, cursor);
    if (token.startsWith('$')) {
        return {
            mode: 'path' as const,
            start: tokenStart,
            end: cursor,
            query: token,
        };
    }

    return null;
}

function coercePrimitiveValue(rawValue: string, originalValue: unknown) {
    if (typeof originalValue === 'number') {
        const parsed = Number(rawValue);
        return Number.isNaN(parsed) ? rawValue : parsed;
    }

    if (typeof originalValue === 'boolean') {
        if (rawValue === 'true') return true;
        if (rawValue === 'false') return false;
    }

    return rawValue;
}

function TemplateValueInput({
    value,
    originalValue,
    onValueChange,
    suggestions,
    inputProps,
}: {
    value: string;
    originalValue: unknown;
    onValueChange: (value: any) => void;
    suggestions: TemplateSuggestion[];
    inputProps?: React.ComponentProps<'input'>;
}) {
    const inputRef = useRef<HTMLInputElement | null>(null);
    const [open, setOpen] = useState(false);
    const [query, setQuery] = useState('');
    const [selectedIndex, setSelectedIndex] = useState(0);
    const [contextRange, setContextRange] = useState<{ start: number; end: number; mode: 'placeholder' | 'path' } | null>(null);

    const filteredSuggestions = useMemo(() => {
        const normalized = query.trim().toLowerCase();
        const base = !normalized
            ? suggestions
            : suggestions.filter((item) =>
                item.path.toLowerCase().includes(normalized) || item.label.toLowerCase().includes(normalized),
            );
        return base.slice(0, 40);
    }, [query, suggestions]);

    const syncAutocomplete = React.useCallback((nextValue: string, cursor: number, forceOpen = false) => {
        const context = getTemplateContext(nextValue, cursor);
        if (!context) {
            setOpen(false);
            setQuery('');
            setContextRange(null);
            setSelectedIndex(0);
            return;
        }

        setQuery(context.query);
        setContextRange({ start: context.start, end: context.end, mode: context.mode });
        setSelectedIndex(0);
        setOpen(forceOpen || context.query.length > 0);
    }, []);

    const applySuggestion = React.useCallback((suggestion: TemplateSuggestion) => {
        const input = inputRef.current;
        if (!input) {
            return;
        }

        const selectionStart = input.selectionStart ?? value.length;
        const currentContext = contextRange ?? getTemplateContext(value, selectionStart);
        let nextValue = value;
        let cursorPosition = selectionStart;

        if (currentContext?.mode === 'placeholder') {
            const before = value.slice(0, currentContext.start);
            const after = value.slice(currentContext.end);
            nextValue = `${before}{{${suggestion.path}}}${after}`;
            cursorPosition = before.length + suggestion.path.length + 4;
        } else if (currentContext?.mode === 'path') {
            const before = value.slice(0, currentContext.start);
            const after = value.slice(currentContext.end);
            nextValue = `${before}${suggestion.path}${after}`;
            cursorPosition = before.length + suggestion.path.length;
        } else {
            const before = value.slice(0, selectionStart);
            const after = value.slice(input.selectionEnd ?? selectionStart);
            nextValue = `${before}{{${suggestion.path}}}${after}`;
            cursorPosition = before.length + suggestion.path.length + 4;
        }

        onValueChange(coercePrimitiveValue(nextValue, originalValue));
        setOpen(false);
        setQuery('');
        setContextRange(null);

        requestAnimationFrame(() => {
            input.focus();
            input.setSelectionRange(cursorPosition, cursorPosition);
        });
    }, [contextRange, onValueChange, originalValue, value]);

    const insertTemplateAtCursor = React.useCallback((templatePath: string) => {
        const input = inputRef.current;
        if (!input) {
            onValueChange(coercePrimitiveValue(`{{${templatePath}}}`, originalValue));
            return;
        }

        const start = input.selectionStart ?? value.length;
        const end = input.selectionEnd ?? value.length;
        const before = value.slice(0, start);
        const after = value.slice(end);
        const nextValue = `${before}{{${templatePath}}}${after}`;
        const cursorPosition = before.length + templatePath.length + 4;

        onValueChange(coercePrimitiveValue(nextValue, originalValue));
        requestAnimationFrame(() => {
            input.focus();
            input.setSelectionRange(cursorPosition, cursorPosition);
        });
    }, [onValueChange, originalValue, value]);

    const [{ isOver }, drop] = useDrop(() => ({
        accept: ItemTypes.OUTPUT_VALUE,
        drop: (item: { value: string }) => {
            insertTemplateAtCursor(item.value);
        },
        collect: (monitor) => ({
            isOver: !!monitor.isOver(),
        }),
    }), [insertTemplateAtCursor]);

    return (
        <Popover open={open && filteredSuggestions.length > 0} onOpenChange={setOpen}>
            <div
                ref={drop as any}
                className={cn(
                    'flex-1 min-w-0 rounded-md transition-colors',
                    isOver && 'ring-1 ring-violet-500/50 bg-violet-500/10',
                )}
            >
                <PopoverAnchor asChild>
                    <div>
                        <Input
                            ref={inputRef}
                            value={value}
                            onChange={(e) => {
                                const nextValue = e.target.value;
                                onValueChange(coercePrimitiveValue(nextValue, originalValue));
                                syncAutocomplete(nextValue, e.target.selectionStart ?? nextValue.length);
                            }}
                            onClick={(e) => {
                                syncAutocomplete((e.target as HTMLInputElement).value, (e.target as HTMLInputElement).selectionStart ?? 0);
                            }}
                            onKeyUp={(e) => {
                                const target = e.currentTarget;
                                syncAutocomplete(target.value, target.selectionStart ?? target.value.length);
                            }}
                            onBlur={() => {
                                window.setTimeout(() => setOpen(false), 120);
                            }}
                            onKeyDown={(e) => {
                                if ((e.ctrlKey || e.metaKey) && e.key === ' ') {
                                    e.preventDefault();
                                    syncAutocomplete(value, inputRef.current?.selectionStart ?? value.length, true);
                                    return;
                                }

                                if (!open || filteredSuggestions.length === 0) {
                                    return;
                                }

                                if (e.key === 'ArrowDown') {
                                    e.preventDefault();
                                    setSelectedIndex((current) => (current + 1) % filteredSuggestions.length);
                                    return;
                                }

                                if (e.key === 'ArrowUp') {
                                    e.preventDefault();
                                    setSelectedIndex((current) =>
                                        current === 0 ? filteredSuggestions.length - 1 : current - 1,
                                    );
                                    return;
                                }

                                if (e.key === 'Enter' || e.key === 'Tab') {
                                    e.preventDefault();
                                    applySuggestion(filteredSuggestions[selectedIndex]);
                                    return;
                                }

                                if (e.key === 'Escape') {
                                    e.preventDefault();
                                    setOpen(false);
                                }
                            }}
                            className="h-7 text-[10px] py-1 bg-input border-border text-primary placeholder:text-secondary font-mono focus:ring-violet-500/30"
                            placeholder="Valor ou {{$.path}}"
                            {...inputProps}
                        />
                    </div>
                </PopoverAnchor>
            </div>
            <PopoverContent
                align="start"
                side="bottom"
                className="w-[360px] p-0 border-border bg-zinc-950"
                onOpenAutoFocus={(e) => e.preventDefault()}
            >
                <Command className="bg-transparent">
                    <CommandList>
                        <CommandEmpty>Nenhuma variavel encontrada.</CommandEmpty>
                        <CommandGroup heading="Variaveis disponiveis">
                            {filteredSuggestions.map((suggestion, index) => (
                                <CommandItem
                                    key={suggestion.path}
                                    value={`${suggestion.label} ${suggestion.path}`}
                                    onMouseDown={(e) => e.preventDefault()}
                                    onSelect={() => applySuggestion(suggestion)}
                                    className={cn(
                                        'flex items-center justify-between gap-3 text-primary',
                                        index === selectedIndex && 'bg-violet-500/10',
                                    )}
                                >
                                    <span className="truncate font-mono text-[11px]">
                                        {suggestion.path}
                                    </span>
                                    <span className="shrink-0 text-[9px] uppercase tracking-widest text-zinc-500">
                                        {suggestion.kind}
                                    </span>
                                </CommandItem>
                            ))}
                        </CommandGroup>
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
}

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
    templateContext,
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
    templateContext?: Record<string, unknown>;
}) {
    const [_expandOnce, setExpandOnce] = React.useState(expandOnce)
    const isArray = Array.isArray(data);
    const templateSuggestions = useMemo(
        () => (templateContext ? buildTemplateSuggestions(templateContext) : []),
        [templateContext],
    );

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
                                    templateSuggestions={templateSuggestions}
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
    draggableValue = true,
    templateSuggestions = [],
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
    draggableValue?: boolean;
    templateSuggestions?: TemplateSuggestion[];
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
                    if (item.data && typeof item.data === 'object' && value !== null && value !== undefined && typeof value !== 'string') {
                        onParamChange(parentPath, name, item.data);
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
                ref={onParamChange ? undefined : drop as any}
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
                                     <TemplateValueInput
                                        value={value === null || value === undefined ? "" : String(value)}
                                        originalValue={value}
                                        onValueChange={(nextValue) =>
                                            onParamChange(path.split('.').slice(0, -1).join('.'), name, nextValue)
                                        }
                                        suggestions={templateSuggestions}
                                        inputProps={inputProps}
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
                                    templateSuggestions={templateSuggestions}
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


