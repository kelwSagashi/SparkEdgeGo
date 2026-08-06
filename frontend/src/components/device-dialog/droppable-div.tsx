import React, { useEffect } from "react";
import { useDrop } from "react-dnd";

const DroppableDiv: React.FC<(
    React.ComponentProps<"div"> & {
        accept: string,
        over?: () => void
    })> = React.memo(function DDiv({
        accept, onDrop, over, ...props
    }) {
        const [{ isOver }, drop] = useDrop({
            accept,
            drop: onDrop,
            collect: (monitor) => ({
                isOver: monitor.isOver(),
                canDrop: monitor.canDrop(),
            }),
        });

        useEffect(() => {
            if (isOver) over?.();
        }, [over, isOver])

        return (
            <div
                ref={drop as any}
                {...props}
            />
        );
    });
export default DroppableDiv;
