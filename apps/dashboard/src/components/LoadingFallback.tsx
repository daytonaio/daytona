import React from 'react'

const LoadingFallback = () => (
  <div className="fixed top-0 left-0 w-full h-full p-6 bg-black z-[3]">
    <div className="flex items-center gap-2">
      <div className="w-4 h-4 border-2 border-white border-t-gray-600 rounded-full animate-spin" />
      <p className="text-white text-sm">Loading...</p>
    </div>
  </div>
)

export default LoadingFallback
