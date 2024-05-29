import React from 'react';
import AssemblyLine from './AssemblyLine';

interface AssemblyListingProps {
    currentPC: number;
    lines: Array<{
        hasBreakpoint?: boolean;
        pc: number;
        text: string;
    }>
}

const AssemblyListing: React.FC<AssemblyListingProps> = (props) => {
    return (
        <div className="assembly-listing">
            {props.lines.map((line, index) => (
                <AssemblyLine
                    key={index}
                    hasBreakpoint={line.hasBreakpoint}
                    isSelected={line.pc === props.currentPC}
                    pc={line.pc}
                    text={line.text}
                />
            ))}
        </div>
    );
};

export default AssemblyListing;
