"use client"
import React from 'react'
import { Bell, Search, Terminal } from 'lucide-react'

import {
    Settings
} from "lucide-react"

import {
    CommandDialog,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList
} from "@/components/ui/command"

export function CommandDialogSearch({
    open,
    setOpen,
    search,
    setSearch,
}: {
    open: boolean,
    setOpen: React.Dispatch<React.SetStateAction<boolean>>
    search: string,
    setSearch: React.Dispatch<React.SetStateAction<string>>
}) {

    React.useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === "j" && (e.metaKey || e.ctrlKey)) {
                e.preventDefault()
                setOpen((open) => !open)
            }
        }

        document.addEventListener("keydown", down)
        return () => document.removeEventListener("keydown", down)
    }, [setOpen]);

    return (
        <>
            <span className="absolute right-4 top-1/2 -translate-y-1/2 bg-opacity-50 text-xs px-2 py-1">
                <p className="text-secondary text-sm">
                    <kbd className="text-muted pointer-events-none inline-flex h-5 items-center gap-1 rounded border border-border/30 px-1.5 font-mono text-[10px] font-medium opacity-100 select-none">
                        <span className="text-xs">Ctrl + J</span>
                    </kbd>
                </p>
            </span>
            <CommandDialog className='bg-black' open={open} onOpenChange={setOpen}>
                <CommandInput className='' value={search} placeholder="Digite algo" onValueChange={setSearch}/>
                <CommandList className='border-none'>
                    <CommandEmpty>No results found.</CommandEmpty>
                    <CommandGroup heading="Ações">
                        <CommandItem className='text-primary'>
                            <Settings />
                            <span>Configurações</span>
                        </CommandItem>
                        <CommandItem className='text-primary'>
                            <Bell />
                            <span>Notificações</span>
                        </CommandItem>
                        <CommandItem className='text-primary'>
                            <Terminal />
                            <span>Logs</span>
                        </CommandItem>
                    </CommandGroup>
                </CommandList>
            </CommandDialog>
        </>
    )
}


function SearchGlobal() {
    const [open, setOpen] = React.useState(false)
    const [search, setSearch] = React.useState("");
    return (
        <div className="relative flex-grow max-w-xl mx-auto">
            <Search color='white' className='absolute left-2 top-1/2 -translate-y-1/2' />
            <input
                type="text"
                placeholder={"Pesquisar" + (search && ":") + search}
                onFocus={() => setOpen(true)}
                className="text-primary bg-secondary-foreground w-full pl-10 pr-24 py-2 rounded-md text-sm outline-none focus:border-accentBlue transition-colors duration-200"
            />
            <CommandDialogSearch open={open} setOpen={setOpen} search={search} setSearch={setSearch} />
        </div>
    )
}

export default SearchGlobal
