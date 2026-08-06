import { cn } from "@/lib/utils";
import { Input } from "../ui/input";
import { useDrop } from "react-dnd";
import type { DroppableInputProps } from "./types";
import React from "react";

const DroppableInput: React.FC<DroppableInputProps> = React.memo(function DInput({
    accept, onDrop, className, ...props
}) {
    const [{ isOver, canDrop }, drop] = useDrop({
        accept,
        drop: onDrop,
        collect: (monitor) => ({
            isOver: monitor.isOver(),
            canDrop: monitor.canDrop(),
        }),
    });

    const isActive = isOver && canDrop;
    let backgroundColor = ''
    if (isActive) {
        backgroundColor = 'bg-foreground text-primary'
    } else if (canDrop) {
        backgroundColor = ''
    }
    const borderColor = isActive ? 'ring-1' : (canDrop ? 'border-dashed border-muted' : '');

    return (
        <div
            ref={drop as any}
        >
            <Input
                className={cn(
                    className,
                    borderColor,
                    backgroundColor,
                    "border-border/60 placeholder:text-foreground rounded text-primary transition-all duration-150"
                )}
                {...props}
            />
        </div>
    );
});
export default DroppableInput;
