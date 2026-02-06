/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_CHAIN_ID: string;
  readonly VITE_CHAIN_RPC: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
