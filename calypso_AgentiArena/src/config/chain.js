export const helaTestnet = {
  id: 666888,
  name: 'Hela Official Runtime Testnet',
  network: 'hela-testnet',
  nativeCurrency: {
    name: 'HLUSD',
    symbol: 'HLUSD',
    decimals: 18
  },
  rpcUrls: {
    default: {
      http: ['https://testnet-rpc.helachain.com']
    },
    public: {
      http: ['https://testnet-rpc.helachain.com']
    }
  },
  blockExplorers: {
    default: {
      name: 'HeLa Explorer',
      url: 'https://testnet-scan.helachain.com'
    }
  },
  testnet: true
};
