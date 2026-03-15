import { http, createConfig, createStorage } from 'wagmi';
import { metaMask } from 'wagmi/connectors';
import { helaTestnet } from './chain';

export const wagmiConfig = createConfig({
  chains: [helaTestnet],
  connectors: [metaMask()],
  storage: createStorage({
    storage: window.localStorage,
  }),
  transports: {
    [helaTestnet.id]: http('https://testnet-rpc.helachain.com'),
  },
});
