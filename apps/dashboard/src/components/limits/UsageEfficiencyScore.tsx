import type React from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface EfficiencyScore {
  grade: 'A' | 'B' | 'C' | 'D' | 'F'
  score: number
  factors: {
    utilization: number
    consistency: number
    peakManagement: number
  }
  recommendations: string[]
}

interface UsageEfficiencyScoreProps {
  efficiencyScore: EfficiencyScore
}

export const UsageEfficiencyScore: React.FC<UsageEfficiencyScoreProps> = ({ efficiencyScore }) => {
  const getGradeColor = (grade: string) => {
    switch (grade) {
      case 'A':
        return 'bg-green-500 text-white'
      case 'B':
        return 'bg-blue-500 text-white'
      case 'C':
        return 'bg-yellow-500 text-white'
      case 'D':
        return 'bg-orange-500 text-white'
      case 'F':
        return 'bg-red-500 text-white'
      default:
        return 'bg-gray-500 text-white'
    }
  }

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-5 h-5 rounded bg-muted flex items-center justify-center">
              <div className="w-2 h-2 bg-foreground rounded-full" />
            </div>
            <span>Usage Efficiency Score</span>
          </div>
          <Badge className={`${getGradeColor(efficiencyScore.grade)}`}>{efficiencyScore.grade}</Badge>
        </CardTitle>
        <CardDescription className="text-sm">
          How efficiently you're using your current tier capacity based on utilization, consistency, and peak
          management.
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-0 space-y-4">
        <div className="grid grid-cols-4 gap-3">
          <div className="text-center">
            <div className="text-2xl font-bold">{efficiencyScore.score}</div>
            <div className="text-xs text-muted-foreground">Overall Score</div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">{efficiencyScore.factors.utilization}%</div>
            <div className="text-xs text-muted-foreground">Utilization</div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">{efficiencyScore.factors.consistency}%</div>
            <div className="text-xs text-muted-foreground">Consistency</div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">{efficiencyScore.factors.peakManagement}%</div>
            <div className="text-xs text-muted-foreground">Peak Management</div>
          </div>
        </div>
        {efficiencyScore.recommendations.length > 0 && (
          <div className="p-3 bg-muted/50 rounded-lg">
            <h4 className="font-medium mb-2 text-sm">Recommendations</h4>
            <ul className="space-y-1">
              {efficiencyScore.recommendations.map((rec, index) => (
                <li key={index} className="text-xs text-muted-foreground flex items-center gap-2">
                  <div className="w-1 h-1 bg-muted-foreground rounded-full" />
                  {rec}
                </li>
              ))}
            </ul>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
