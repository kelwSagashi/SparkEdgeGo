/* eslint-disable @typescript-eslint/no-explicit-any */
import { ItemTypes } from '@/lib/constants';
import React from 'react';
import { useDrag } from 'react-dnd';

interface DraggableValueProps {
    value: string;
    type: string,
    isDropped: boolean;
    data?: any;
    children?: React.ReactNode;
}

const DraggableValue: React.FC<DraggableValueProps> = React.memo(function Box({ value, type, isDropped, data, children }) {
    const [{ isDragging }, drag] = useDrag(() => ({
        type: type,
        item: { value, data }, 
        collect: (monitor) => ({
            isDragging: monitor.isDragging(),
        }),
    }), [value, type, data]);

    return (
        <div
            ref={drag as any}
            className={`cursor-grab ${isDragging ? 'opacity-50 border border-primary ring-2 ring-primary' : 'hover:bg-accent'} 
                       border-none break-all w-full text-sm py-1 px-2 rounded transition-all duration-150`}
            style={{ opacity: isDragging ? 0.5 : 1 }}
            data-testid={`output_value`}
        >
            {children || value}
        </div>
    );
});

export default DraggableValue;
