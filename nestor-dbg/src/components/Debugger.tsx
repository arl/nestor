import React from 'react';
import { CPUState } from '../types';
import useWS from '../ws/hook';

export interface DebuggerState {
  cpu: CPUState;
}

// Define ButtonProps interface
interface ButtonProps {
  onClick: () => void;
  disabled: boolean;
  children: React.ReactNode;
}

// Button component
const Button: React.FC<ButtonProps> = ({ onClick, disabled, children }) => (
  <button onClick={onClick} disabled={disabled}>
    {children}
  </button>
);

// Debugger component
function Debugger() {
  const [debuggerState, setDebuggerState] =
    React.useState<DebuggerState | null>(null);

  const [ws] = useWS();

  React.useEffect(() => {
    if (!ws) return;

    return ws.on('state', (data) => setDebuggerState(data));
  }, [ws]);

  const handleStart = () => {
    /*setDebuggerState(CPUState.Running);*/
  };
  const handlePause = () => {
    /*setDebuggerState(CPUState.Paused);*/
    ws?.send('state', { cpu: CPUState.Paused });
  };
  const handleStep = () => {
    /*setDebuggerState(CPUState.Stepping);*/
  };

  return (
    <div>
      <Button
        onClick={handleStart}
        disabled={debuggerState?.cpu === CPUState.Running}
      >
        Start
      </Button>
      <Button
        onClick={handlePause}
        disabled={debuggerState?.cpu !== CPUState.Running}
      >
        Pause
      </Button>
      <Button
        onClick={handleStep}
        disabled={debuggerState?.cpu === CPUState.Running}
      >
        Step
      </Button>
      <div>
        <p>Debugger state: {debuggerState?.cpu}</p>
      </div>
    </div>
  );
}

export default Debugger;
