/// <reference types="react-scripts" />

interface ProcessEnv {
  readonly REACT_APP_API_URL?: string;
}

declare var process: {
  env: ProcessEnv;
};
