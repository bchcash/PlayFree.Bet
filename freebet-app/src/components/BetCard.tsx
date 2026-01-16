import { format } from "date-fns";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Calendar, Clock, Target, X, CircleHelp } from "lucide-react";
import { cn } from "@/lib/utils";
import { ReactNode } from "react";

interface BetCardProps {
  bet: {
    bet_id: string;
    created_at: string;
    home_team: string;
    away_team: string;
    commence_time?: string | undefined;
    bet_type: string;
    bet_amount: number;
    odds: number;
    potential_win: number;
    status: string;
  };
  index: number;
  totalBets: number;
}

export function BetCard({ bet, index, totalBets }: BetCardProps) {
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case "won":
        return "bg-green-500/10 text-green-500 border-green-500/20";
      case "lost":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      case "pending":
        return "bg-yellow-500/10 text-yellow-500 border-yellow-500/20";
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    }
  };

  const getStatusIcon = (status: string): ReactNode => {
    switch (status.toLowerCase()) {
      case "won":
        return <Target className="size-4" />;
      case "lost":
        return <X className="size-4" />;
      case "pending":
        return <Clock className="size-4" />;
      default:
        return <CircleHelp className="size-4" />;
    }
  };

  return (
    <Card className="p-6 hover:shadow-lg transition-all duration-300 border-border/50 bg-gradient-to-br from-card to-card/80">
      <div className="flex flex-col space-y-4">
        {/* Header with number and date */}
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary/10 text-primary font-bold text-sm">
              {index + 1}
            </div>
            <div className="flex flex-col text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <Calendar className="size-4" />
                {format(new Date(bet.created_at), "MMM d, yyyy")}
              </div>
              <div className="flex items-center gap-2 mt-1">
                <Clock className="size-4" />
                {format(new Date(bet.created_at), "HH:mm")}
              </div>
            </div>
          </div>
          <Badge className={cn("text-xs font-bold px-3 py-1", getStatusColor(bet.status))}>
            <div className="flex items-center gap-1">
              {getStatusIcon(bet.status)}
              <span>{bet.status}</span>
            </div>
          </Badge>
        </div>

        {/* Match info */}
        <div className="text-center py-2">
          <div className="font-bold text-lg text-foreground mb-1">
            {bet.home_team} vs {bet.away_team}
          </div>
          {bet.commence_time && (
            <div className="text-sm text-muted-foreground flex items-center justify-center gap-1">
              <Clock className="size-3" />
              {format(new Date(bet.commence_time), "MMM d, yyyy HH:mm")}
            </div>
          )}
        </div>

        {/* Bet details */}
        <div className="flex flex-row items-center justify-between gap-4 pt-2 border-t border-border/30">
          <div className="text-center">
            <div className="text-xs text-muted-foreground uppercase tracking-wider font-bold mb-1">Pick</div>
            <Badge variant="outline" className="bg-purple-500/10 text-purple-600 border-purple-500/20 font-bold">
              {bet.bet_type}
            </Badge>
          </div>

          <div className="text-center">
            <div className="text-xs text-muted-foreground uppercase tracking-wider font-bold mb-1">Stake</div>
            <div className="flex items-center justify-center gap-1 font-bold text-foreground">
              {Number(bet.bet_amount || 0).toLocaleString()}
            </div>
          </div>

          <div className="text-center">
            <div className="text-xs text-muted-foreground uppercase tracking-wider font-bold mb-1">Odds</div>
            <div className="font-bold text-lg text-primary">
              {Number(bet.odds || 0).toFixed(2)}
            </div>
          </div>

          <div className="text-center">
            <div className="text-xs text-muted-foreground uppercase tracking-wider font-bold mb-1">Potential</div>
            <div className="flex items-center justify-center gap-1 font-bold text-green-500">
              {Number(bet.potential_win || 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}
