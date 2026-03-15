import { useReadContract, useWriteContract } from 'wagmi';
import { CONTRACT_ADDRESSES, ABIS } from '../addresses';

export function useArenaEngine() {
  const { data: openArenas, isLoading: isLoadingArenas } = useReadContract({
    address: CONTRACT_ADDRESSES.arena,
    abi: ABIS.ArenaEngine,
    functionName: 'getOpenArenas',
  });

  const { writeContract: createArena } = useWriteContract();

  return {
    openArenas,
    isLoadingArenas,
    createArena: (args) => createArena({
      address: CONTRACT_ADDRESSES.arena,
      abi: ABIS.ArenaEngine,
      functionName: 'createArena',
      args,
    }),
  };
}
