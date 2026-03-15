import { useWriteContract } from 'wagmi';
import { useWallet } from '../../hooks/useWallet';
import { CONTRACT_ADDRESSES, ABIS } from '../addresses';

export function useAgentVault() {
  const { address } = useWallet();

  const { writeContract: payForTask } = useWriteContract();
  const { writeContract: depositStake } = useWriteContract();

  return {
    payForTask: (args) => payForTask({
      address: CONTRACT_ADDRESSES.vault,
      abi: ABIS.AgentVault,
      functionName: 'payForTask',
      args,
    }),
    depositStake: (args) => depositStake({
      address: CONTRACT_ADDRESSES.vault,
      abi: ABIS.AgentVault,
      functionName: 'depositStake',
      args,
    }),
  };
}
