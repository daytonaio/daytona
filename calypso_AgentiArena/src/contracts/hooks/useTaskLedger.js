import { useReadContract, useWriteContract } from 'wagmi';
import { CONTRACT_ADDRESSES, ABIS } from '../addresses';
import { useWallet } from '../../hooks/useWallet';

export function useTaskLedger() {
  const { address } = useWallet();

  const { data: userTasks, isLoading: isLoadingTasks } = useReadContract({
    address: CONTRACT_ADDRESSES.taskLedger,
    abi: ABIS.TaskLedger,
    functionName: 'getUserTasks',
    args: [address],
    query: { enabled: !!address },
  });

  const { writeContract: createTask } = useWriteContract();

  return {
    userTasks,
    isLoadingTasks,
    createTask: (args) => createTask({
      address: CONTRACT_ADDRESSES.taskLedger,
      abi: ABIS.TaskLedger,
      functionName: 'createTask',
      args,
    }),
  };
}
