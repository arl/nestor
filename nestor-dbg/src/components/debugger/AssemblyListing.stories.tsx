import AssemblyListing from './AssemblyListing';

export default {
    component: AssemblyListing,
    title: 'AssemblyListing',
    tags: ['autodocs'],
};

export const Default = {
    args: {
        currentPC: 0xD173,
        lines: [
            { pc: 0xD168, text: 'EOR #$FF', hasBreakpoint: false },
            { pc: 0xD169, text: 'AND keydown', hasBreakpoint: true },
            { pc: 0xD170, text: 'BEQ $D172', hasBreakpoint: false },
            { pc: 0xD173, text: 'RTS', hasBreakpoint: true },
            { pc: 0xD174, text: 'EOR #$23', hasBreakpoint: false },
        ],
    },
};
