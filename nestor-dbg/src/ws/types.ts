import { CPUStatus } from '../types';

// Types for the responses from the emulator to the debugger.
export interface EmulatorStateResponse {
  status: CPUStatus;
  pc: number;
}

export interface WSResponse {
  event: string;
  data: EmulatorStateResponse /* | other response types */;
}

// Types for the requests by the debugger to the emulator.
export type SetCPUStateRequest = string;

export interface WSRequest {
  event: string;
  data: SetCPUStateRequest /* | other request types */;
}