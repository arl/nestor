import AssemblyLine from './AssemblyLine';

export default {
    component: AssemblyLine,
    title: 'AssemblyLine',
    tags: ['autodocs'],
};

export const Default = {
    args: {
        hasBreakpoint: false,
        isSelected: false,
        pc: 0xC000,
        text: 'STY $2000',
    },
};

export const CurrentPC = {
    args: {
        hasBreakpoint: false,
        isSelected: true,
        pc: 0xC000,
        text: 'STY $2000',
    },
};

export const HasBreakpoint = {
    args: {
        hasBreakpoint: true,
        isSelected: false,
        pc: 0xC000,
        text: 'STY $2000',
    },
};

export const CurrentBreakpoint = {
    args: {
        hasBreakpoint: true,
        isSelected: true,
        pc: 0xC000,
        text: 'STY $2000',
    },
};
