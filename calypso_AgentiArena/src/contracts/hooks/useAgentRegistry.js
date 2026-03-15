import { useReadContract, useWriteContract } from 'wagmi';
import { useWallet } from '../../hooks/useWallet';
import { CONTRACT_ADDRESSES, ABIS } from '../addresses';
import { useState, useEffect } from 'react';
import { agents as mockAgents } from '../../data/agents';

export function useAgentRegistry() {
  const { address } = useWallet();

  const { data: allAgents, isLoading: isContractLoading } = useReadContract({
    address: CONTRACT_ADDRESSES.registry,
    abi: ABIS.AgentRegistry,
    functionName: 'getAllAgents',
  });

  const [agentsToReturn, setAgentsToReturn] = useState(mockAgents);
  const [isLoadingAgents, setIsLoadingAgents] = useState(true);

  useEffect(() => {
    // If contract returns data and it's not empty, use it
    if (allAgents && allAgents.length > 0) {
      setAgentsToReturn(allAgents);
      setIsLoadingAgents(false);
      return;
    }

    // Otherwise, set a timeout to fallback to mock data
    const timer = setTimeout(() => {
      console.log('Contract read timeout or empty, falling back to mock agents');
      setAgentsToReturn(mockAgents);
      setIsLoadingAgents(false);
    }, 3000);

    return () => clearTimeout(timer);
  }, [allAgents]);

  // Log warning if environment variables or addresses are missing
  useEffect(() => {
    if (!CONTRACT_ADDRESSES.registry || CONTRACT_ADDRESSES.registry === '0x0000000000000000000000000000000000000000') {
      console.warn("WARNING: AgentRegistry address is missing or set to zero! Check addresses.js or deploy scripts. Falling back to mock data.");
    }
  }, []);

  const { writeContract: registerAgent } = useWriteContract();

  return {
    allAgents: agentsToReturn,
    isLoadingAgents,
    registerAgent: (args) => registerAgent({
      address: CONTRACT_ADDRESSES.registry,
      abi: ABIS.AgentRegistry,
      functionName: 'registerAgent',
      args,
    }),
  };
}
