import React from 'react';
import './AssemblyLine.css';

interface AssemblyLineProps {
    hasBreakpoint?: boolean;
    isSelected?: boolean;
    pc: number;
    text: string;
    onSelect?: () => void;
    onSetBreakpoint?: () => void;
}

const AssemblyLine: React.FC<AssemblyLineProps> = (props) => {
    return (
        <div className={`assembly-line ${props.isSelected ? 'selected' : ''} ${props.hasBreakpoint ? 'breakpoint' : ''}`}>
            <div className="line-content">
                <span className="pc">{props.pc.toString(16).toUpperCase()}</span>
                <span className="text">{props.text}</span>
            </div>
        </div>
    );
};

export default AssemblyLine;