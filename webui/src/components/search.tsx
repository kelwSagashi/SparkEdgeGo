import React from 'react'
import {Search as SearchIcon} from 'lucide-react'

function Search({
    place_holder,
    value,
    onChange
}: {
    place_holder?: string,
    value: string,
    onChange: React.ChangeEventHandler<HTMLInputElement>
    }) {
    return (
        <div className="grow max-w-xl mx-auto border-2 rounded-sm border-primary-foreground/40">
            <div className='relative'>
            <SearchIcon color='white' className='absolute left-2 top-1.5'/>
            </div>
            <input
                type="text"
                value={value}
                onChange={onChange}
                placeholder={place_holder ?? "Pesquisar"}
                className="text-secondary w-full pl-10 pr-24 py-2 rounded-md text-sm outline-none focus:border-accentBlue transition-colors duration-200"
            />
        </div>
    )
}

export default Search
