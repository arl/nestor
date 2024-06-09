import React from 'react';
import { CPUStatus } from '../types';
import { EmulatorStateResponse } from '../ws/types';
import useWS from '../ws/hook';

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
    React.useState<EmulatorStateResponse | null>(null);

  const [ws] = useWS();

  React.useEffect(() => {
    if (!ws) return;

    return ws.on('state', (resp) => {
      console.log('Debugger received state response', resp);
      setDebuggerState(resp)
    });
  }, [ws]);

  const handleStart = () => {
    ws?.send({ event: 'set-cpu-state', data: "run" });
  }
  const handlePause = () => {
    ws?.send({ event: 'set-cpu-state', data: "pause" });
  };
  const handleStep = () => {
    ws?.send({ event: 'set-cpu-state', data: "step" });
  };

  return (
    <div>
      <Button
        onClick={handleStart}
        disabled={debuggerState?.status === CPUStatus.Running}
      >
        Start
      </Button>
      <Button
        onClick={handlePause}
        disabled={debuggerState?.status !== CPUStatus.Running}
      >
        Pause
      </Button>
      <Button
        onClick={handleStep}
        disabled={debuggerState?.status === CPUStatus.Running}
      >
        Step
      </Button>
      <div>
        <p>Debugger state: {debuggerState?.status}</p>
        <p>Current PC: 0x{debuggerState?.pc.toString(16)}</p>
      </div>
    </div>
  );
}

export default Debugger;
